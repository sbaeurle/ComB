from typing import Optional, List, Dict
import numpy as np
import cv2
import argparse
from io import BytesIO

import asyncio
from grpclib.utils import graceful_exit
from grpclib.server import Server
import tracking

import logging
import time
import requests
import json
import helper

import cProfile
import pstats

def display_image(objects: dict, image: np.ndarray):
    font = cv2.FONT_HERSHEY_PLAIN
    for _, object in objects.items():
        x, y, w, h = object.bbox
        label = str(object.id)
        cv2.rectangle(image, (x, y), (x+w, y+h), object.color, 2)
        cv2.putText(image, label, (x, y - 5), font, 1, object.color, 1)
    cv2.imshow("Object Tracker", image)

    if cv2.waitKey(1) & 0xFF == ord('q'):
        cv2.destroyAllWindows()
        exit()


def _new_tracker(name):
    trackers = {
        "csrt": cv2.legacy.TrackerCSRT_create,
        "kcf": cv2.legacy.TrackerKCF_create,
        "boosting": cv2.legacy.TrackerBoosting_create,
        "mil": cv2.legacy.TrackerMIL_create,
        "tld": cv2.legacy.TrackerTLD_create,
        "medianflow": cv2.legacy.TrackerMedianFlow_create,
        "mosse": cv2.legacy.TrackerMOSSE_create
    }
    return trackers[name]()


class TrackedObject():
    def __init__(self, tracker, bbox, conf, clazz, id):
        self.tracker = tracker
        self.bbox = bbox
        self.conf = conf
        self.clazz = clazz
        self.id = id
        self.disappeared_count = 0
        self.color = np.random.uniform(0, 255, size=3)


class Tracker(tracking.ObjectTrackingBase):
    def __init__(self, metric_service: helper.MetricService, result_service: helper.MetricService):
        self.objects: dict[int, TrackedObject] = dict()
        self.initialized = False
        self.metric_service = metric_service
        self.result_service = result_service
        self.last_id = 1

    def _add_new_object(self, img, detection: tracking.DetectionsDetection, id: int):
        tracker = _new_tracker(ARGS["tracker"])
        tracker.init(img, detection.bbox)
        self.objects[id] = TrackedObject(tracker, np.asarray(
            detection.bbox), detection.conf, detection.clazz, id)

        logging.debug(
            f'Added {detection.clazz} with id {id} at {detection.bbox}')

    async def add_detections(
        self, frame: tracking.Frame, detections: Optional[List[tracking.DetectionsDetection]]
    ) -> tracking.Empty:
        img = np.load(BytesIO(frame.image), allow_pickle=False)
        img = cv2.imdecode(img, cv2.IMREAD_UNCHANGED)

        # await self.track_objects(frame.image, frame.timestamp, frame.frame_number)

        if not self.initialized:
            self.initialized = True
            detections = detections
        else:
            if len(self.objects) > 0 and len(detections) > 0:
                indices: List[int] = list()
                matches: dict[int, int] = dict()
                for i, det in enumerate(detections):
                    iou = 0.0
                    for id2, obj in self.objects.items():
                        try:
                            bb1 = {'x1': det.bbox[0], 'x2': det.bbox[0] + det.bbox[2],
                                   'y1': det.bbox[1], 'y2': det.bbox[1] + det.bbox[3]}
                            bb2 = {'x1': obj.bbox[0], 'x2': obj.bbox[0] + obj.bbox[2],
                                   'y1': obj.bbox[1], 'y2': obj.bbox[1] + obj.bbox[3]}
                            iou2 = helper.get_iou(bb1, bb2)
                        finally:
                            if iou2 > ARGS["confidencethreshold"] and iou < iou2:
                                iou = iou2
                                matches[id2] = i
                    if iou == 0.0:
                        indices.append(i)
                
                for id, i in matches.items():
                    tracker = _new_tracker(ARGS["tracker"])
                    tracker.init(img, detections[i].bbox)
                    self.objects[id].tracker = tracker

                detections = [x for i, x in enumerate(
                    detections) if i in indices]

        for detection in detections:
            self._add_new_object(img, detection, self.last_id)
            self.last_id += 1

        if ARGS["visual"]:
            display_image(self.objects, img)

        return tracking.Empty()

    async def track_objects(self, image: bytes, timestamp: int, frame_number: int) -> tracking.Empty:
        img = np.load(BytesIO(image), allow_pickle=False)
        img = cv2.imdecode(img, cv2.IMREAD_UNCHANGED)

        with helper.Timer("processing-time", self.metric_service):
            for i in list(self.objects.keys()):
                success, bbox = self.objects[i].tracker.update(img)

                if success:
                    logging.debug(
                        f'Updating {self.objects[i].clazz} with id {self.objects[i].id} from {self.objects[i].bbox} to {bbox}')
                    self.objects[i].bbox = np.asarray(bbox, dtype=int)
                else:
                    self.objects[i].disappeared_count += 1
                    if self.objects[i].disappeared_count >= ARGS["deregister_threshold"]:
                        tmp = self.objects.pop(self.objects[i].id)
                        logging.debug(
                            f'Deregister {tmp.clazz} with id {tmp.id} at {tmp.bbox}')

        self.metric_service.SetMetrics(
            {"tracked-objects": len(self.objects), "frame-number": frame_number})

        if ARGS["visual"]:
            display_image(self.objects, img)

        asyncio.create_task(self.metric_service.PushMetrics())

        with helper.Timer("result-time", self.metric_service):
            detections = list()
            for i, object in self.objects.items():
                detections.append({
                    "id": object.id,
                    "class": object.clazz,
                    "conf": object.conf,
                    "bb_left": int((object.bbox[0]/ARGS["image_size"][0])*ARGS["dataset_size"][0]),
                    "bb_width": int((object.bbox[2]/ARGS["image_size"][0])*ARGS["dataset_size"][0]),
                    "bb_top": int((object.bbox[1]/ARGS["image_size"][1])*ARGS["dataset_size"][1]),
                    "bb_height": int((object.bbox[3]/ARGS["image_size"][1])*ARGS["dataset_size"][1])
                })

            self.result_service.SetMetrics({
                "count": frame_number,
                "detections": detections
            })

            asyncio.create_task(self.result_service.PushMetrics())

        return tracking.Empty()


async def main():
    if not ARGS["evaluation_address"] is None:
        metric_service = helper.MetricService(
            ARGS["evaluation_address"] + "/tracking")
        result_service = helper.MetricService(
            ARGS["evaluation_address"] + "/pipeline-results")
    else:
        metric_service = helper.MetricService()
        result_service = helper.MetricService()

    tracker = Tracker(metric_service, result_service)

    server = Server([tracker])
    with graceful_exit([server]):
        await server.start(port=8181)
        await server.wait_closed()


if __name__ == "__main__":
    # Setup CLI Arguments
    ap = argparse.ArgumentParser()
    ap.add_argument("--tracker", type=str, default="kcf",
                    help="OpenCV object tracker type")
    ap.add_argument("--log-level", type=str, default="INFO",
                    help="Verbosity of generate application logs")
    ap.add_argument("--image-size", type=int, nargs=2, default=[320, 320])
    ap.add_argument("--dataset-size", type=int, nargs=2, default=[1920, 1080])
    ap.add_argument("--confidencethreshold", type=float, default=0.5,
                    help="minimum iou for two objects to be the same")
    ap.add_argument("--visual", action='store_true',
                    help="""display images graphically, needs precompiled support for the used UI framework
                    (see https://docs.opencv.org/master/d7/dfc/group__highgui.html)""")
    ap.add_argument("--deregister-threshold", type=int, default=10,
                    help="number of consecutive misses until object is deregistered")
    ap.add_argument("--evaluation-address", type=str,
                    help="Address and path of evaluation service")
    ap.add_argument("--profiling", action='store_true',
                    help="enable cProfile and export function profiles")
    ARGS = vars(ap.parse_args())

    # Setup Logging
    logging.basicConfig(level=getattr(logging, ARGS["log_level"].upper()))
    logger = logging.getLogger('hpack.hpack').setLevel(logging.ERROR)
    logger = logging.getLogger(
        'urllib3.connectionpool').setLevel(logging.ERROR)

    if ARGS["profiling"]:
        with cProfile.Profile() as pr:
            pr.runcall(asyncio.run, main())
            pr.create_stats()
            stats = pstats.Stats(pr)
            stats.strip_dirs()
            stats.sort_stats(pstats.SortKey.TIME)
            stats.print_stats(10)
    else:
        asyncio.run(main())
