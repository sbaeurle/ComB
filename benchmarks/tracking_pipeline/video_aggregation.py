import argparse
import asyncio
import logging
import queue
import threading
import time
from datetime import datetime
from io import BytesIO

import cv2
import numpy as np
from grpclib.client import Channel

import detection
import helper
import tracking

import cProfile
import pstats


class Ingestion():
    def __init__(self, config: dict, queue: queue.Queue, metric_service: helper.MetricService):
        self.config = config
        self.queue = queue
        # TODO: Evaluate semaphore as other locking mechanism
        self.task = None
        self.metric_service = metric_service

        self.networking = threading.Thread(target=self.serve, daemon=True)
        self.networking.start()

    def serve(self):
        self.loop = asyncio.new_event_loop()
        asyncio.set_event_loop(self.loop)

        detector_address = self.config["detector_address"].split(':')
        tracker_address = self.config["tracker_address"].split(':')
        self.detector_channel = Channel(
            host=detector_address[0], port=detector_address[1])
        self.tracker_channel = Channel(
            host=tracker_address[0], port=tracker_address[1])

        self.detector_stub = detection.ObjectDetectionStub(
            self.detector_channel, timeout=10.0)
        self.tracker_stub = tracking.ObjectTrackingStub(
            self.tracker_channel, timeout=10.0)

        self.loop.run_forever()

    def stop(self):
        self.loop.stop()

    def encode_frame(self, frame):
        with helper.Timer("encoding-time", self.metric_service):
            _, frame = cv2.imencode(".jpg", frame)

            with BytesIO() as encoded_img:
                np.save(encoded_img, frame, allow_pickle=False)
                return encoded_img.getvalue()

    def send_frame_detection(self, frame, frame_number):
        if self.task == None or self.task.done():
            logging.debug(f'Send frame to detection.')
            self.task = asyncio.run_coroutine_threadsafe(loop=self.loop, coro=self.detector_stub.detect_objects(
                image=self.encode_frame(frame), timestamp=int(datetime.utcnow().timestamp()), frame_number=frame_number))
        else:
            self.metric_service.Count("skipped-frames")
            logging.debug(f'Skipping detection frame.')

        asyncio.run_coroutine_threadsafe(
            loop=self.loop, coro=self.metric_service.PushMetrics())

    def send_frame_tracking(self, frame, frame_number):
        if self.task == None or self.task.done():
            logging.debug(f'Send frame to tracking.')
            self.task = asyncio.run_coroutine_threadsafe(loop=self.loop, coro=self.tracker_stub.track_objects(
                image=self.encode_frame(frame), timestamp=int(datetime.utcnow().timestamp()), frame_number=frame_number))
        else:
            self.metric_service.Count("skipped-frames")
            logging.debug(f'Skipping tracking frame.')

        asyncio.run_coroutine_threadsafe(
            loop=self.loop, coro=self.metric_service.PushMetrics())

    def close_channels(self):
        self.detector_channel.close()
        self.tracker_channel.close()


class VideoInput():
    def __init__(self, config: dict(), metric_service: helper.MetricService):
        self.config = config
        self.queue = queue.Queue(self.config["queue_size"])
        self.ingestion = Ingestion(config, self.queue, metric_service)
        self.frame_count = -1
        self.empty_frames = 0
        self.metric_service = metric_service

        self.open_source()

        self.source = threading.Thread(target=self.read_frames, daemon=True)
        self.source.start()

        time.sleep(1)

        self.worker = threading.Thread(target=self.work, daemon=True)
        self.worker.start()

    def open_source(self):
        source = self.config["video_source"]
        logging.debug(f"Opening Video Source {source}")
        self.capture = cv2.VideoCapture()
        self.status = self.capture.open(source)
        if self.status:
            logging.debug(f"Opened {source} successfully.")
        else:
            logging.error(f"Error opening {source}!")

    def work(self):
        self.delay = 1 / self.config["frame_rate"]

        while self.queue.qsize() > 0:
            frame = self.queue.get()
            timer = helper.Timer(metric_name="processing-time",
                                 metric_service=self.metric_service)
            self.frame_count += 1
            self.metric_service.Count("frame-number")

            with timer:
                if (self.frame_count % self.config["detection_frequency"]) == 0:
                    self.ingestion.send_frame_detection(
                        frame[0], self.frame_count)
                else:
                    self.ingestion.send_frame_tracking(
                        frame[0], self.frame_count)

            sleep_duration = self.delay - (timer.time/1000)
            if sleep_duration > 0:
                time.sleep(sleep_duration)

        self.ingestion.stop()

    def read_frames(self):
        while self.status:
            ret, frame = self.capture.read()

            if ret:
                self.empty_frames = 0
                timer = helper.Timer()

                with timer:
                    frame = cv2.resize(frame, tuple(self.config["image_size"]))

                if self.config["grpc"]:
                    logging.debug("Add frame to queue")
                    self.queue.put(item=(frame, timer.time))

                if self.config["visual"]:
                    cv2.imshow("Video Aggregation", frame)
                    cv2.waitKey(1)

            else:
                logging.debug(
                    f'Received empty frame from {self.config["video_source"]}.')
                self.empty_frames += 1

                if self.empty_frames > 10:
                    self.close_stream()

    def close_stream(self):
        self.status = False
        while self.queue.qsize() > 0:
            time.sleep(1)
        logging.info(f'Closing Video Source {self.config["video_source"]}')
        self.ingestion.close_channels()
        self.capture.release()


def _setup_cli_args():
    ap = argparse.ArgumentParser()
    ap.add_argument("--visual", action='store_true',
                    help="display images graphically")
    ap.add_argument("--grpc", action='store_true',
                    help="enable grpc communication")
    ap.add_argument("--detection-frequency", default=20, type=int,
                    help="frequency at which detection service gets called")
    ap.add_argument("--video-source", required=True, type=str,
                    help="filepath or network address of video source")
    ap.add_argument("--tracker-address", required=True, type=str,
                    help="address of tracker")
    ap.add_argument("--detector-address", required=True, type=str,
                    help="address of detector")
    ap.add_argument("--frame-rate", type=int, default=30,
                    help="frame rate of the video")
    ap.add_argument("--queue-size", type=int, default=120,
                    help="size of the frame buffer")
    ap.add_argument("--image-size", type=int, nargs=2, default=[320, 320])
    ap.add_argument("--log-level", type=str, default="INFO",
                    help="verbosity of application logs")
    ap.add_argument("--evaluation-address", type=str,
                    help="url of evaluation service. enables periodic metrics.")
    ap.add_argument("--profiling", action='store_true',
                    help="enable cProfile and export function profiles")
    return vars(ap.parse_args())

def run():
    INPUTS = dict()

    # Setup Logging
    # TODO: Improve Logging and Logging configuration
    logging.basicConfig(level=getattr(logging, ARGS["log_level"].upper()))
    logging.getLogger('hpack.hpack').setLevel(logging.ERROR)
    logging.getLogger('urllib3.connectionpool').setLevel(logging.ERROR)
    
    if not ARGS["evaluation_address"] is None:
        metric_service = helper.MetricService(
            ARGS["evaluation_address"] + "/aggregation")
    else:
        metric_service = helper.MetricService()

    # TODO: Load multiple video inputs
    # TODO: implement async handling of multiple input/output sources
    INPUTS[ARGS["video_source"]] = VideoInput(ARGS, metric_service)

    to_remove = list()
    while len(INPUTS) > 0:
        for key, input in INPUTS.items():
            if not INPUTS[key].status:
                to_remove.append(key)

        for key in to_remove:
            INPUTS.pop(key)
            logging.info(f'Removed input source {key}.')

    time.sleep(10)
    logging.info("No input sources left. Exiting Application")

    if ARGS["visual"]:
        cv2.destroyAllWindows()

if __name__ == "__main__":
    # Setup CLI Arguments
    ARGS = _setup_cli_args()

    if ARGS["profiling"]:
        with cProfile.Profile() as pr:
            pr.runcall(run)
            pr.create_stats()
            stats = pstats.Stats(pr)
            stats.strip_dirs()
            stats.sort_stats(pstats.SortKey.TIME)
            stats.print_stats(10)
    else:
        run()