import typing
import numpy as np
import cv2
import argparse

from io import BytesIO
import asyncio
from grpclib.client import Channel
from grpclib.utils import graceful_exit
from grpclib.server import Server
import tracking
import detection

import logging
import helper

import cProfile
import pstats

class NeuralNet():
    def __init__(self, model_path: str, config_path: str, classes_path: str, conf_threshold: float, nms_threshold: float) -> None:
        self.net, self.classes = self.load_yolo(model_path, config_path, classes_path)

        if "yolox" in model_path:
            raise NotImplementedError("ONNX based YOLOX is not implemented yet!")
        elif "yolov3" in model_path:
            self.model = "yolov3"

        self.conf_threshold = conf_threshold
        self.nms_threshold = nms_threshold

    def load_yolo(self, weights: str, config: str, names: str) -> typing.Tuple[cv2.dnn_Net, list]:
        """Loads pretrained YOLO DNN

        Args:
            weights (str): file path to dnn weights
            config (str): file path to config file
            names (str): filepath to list of classes (coco.names)

        Returns:
            (cv2.dnn_Net, list): Tuple including loaded net and list of classes its trained for
        """
        net = cv2.dnn.readNet(weights, config)

        if(cv2.cuda.getCudaEnabledDeviceCount() > 0):
            net.setPreferableBackend(cv2.dnn.DNN_BACKEND_CUDA)
            net.setPreferableTarget(cv2.dnn.DNN_TARGET_CUDA)           
        elif (cv2.ocl.haveOpenCL()):
            net.setPreferableBackend(cv2.dnn.DNN_BACKEND_OPENCV)
            net.setPreferableTarget(cv2.dnn.DNN_TARGET_OPENCL)
        
        classes = []
        with open(names, "r") as f:
            classes = [line.strip() for line in f.readlines()]
        return net, classes

    def detect_objects(self, image: np.ndarray):
        """Generates blob from provided image, set as input of YOLO and forward through net

        Args:
            image (numpy.ndarray): Loaded Image

        Returns:
            list: list of detections
        """
        blob = cv2.dnn.blobFromImage(image, scalefactor=0.00392, size=image.shape[:2], mean=(0, 0, 0), swapRB=True, crop=False)
        self.net.setInput(blob)
        detections = self.net.forward(self.net.getUnconnectedOutLayersNames())
        return detections

    # TODO: Improve (especially for performance reasons)
    def calculate_box_dimensions_yolov3(self, outputs, h: int, w: int):
        """Selects class with highest confidence, calculates bounding box coordinates and performs non maximum suppression to select best bounding box.

        Args:
            outputs (list): outputs of dnn
            classes (list): list of possible class names
            conf_threshold (float): confidence threshold applied to bounding boxes
            nms_threshold (float): non maximum suppression threshold
            h (int): height of image
            w (int): width of image

        Returns:
            (list (list, float, str)): list of 3-tuples containing coordinates, confidence and class of bounding box
        """
        boxes = []
        confs = []
        detections = []
        for output in outputs:
            for detection in output:
                class_scores = detection[5:]
                classID = np.argmax(class_scores)
                confidence = class_scores[classID]
                if confidence > self.conf_threshold and self.classes[classID] == "person":
                    box = detection[:4] * np.array([w, h, w, h])
                    (centerX, centerY, width, height) = box.astype("int")
                    x = int(centerX - (width / 2))
                    y = int(centerY - (height / 2))
                    box = [x, y, int(width), int(height)]
                    boxes.append(box)
                    confs.append(float(confidence))
                    detections.append((box, confidence, self.classes[classID]))
        indices = cv2.dnn.NMSBoxes(boxes, confs, self.conf_threshold, self.nms_threshold)
        return [x for i, x in enumerate(detections) if i in indices]

class Detector(detection.ObjectDetectionBase):
    def __init__(self, channel: Channel, dnn: NeuralNet, metric_service: helper.MetricService) -> None:
        self.stub = tracking.ObjectTrackingStub(channel)
        self.dnn = dnn
        self.metric_service = metric_service

    async def detect_objects(self, image: bytes, timestamp: int, frame_number: int):
        with helper.Timer("processing-time", self.metric_service):
            img = np.load(BytesIO(image), allow_pickle=False)
            img = cv2.imdecode(img, cv2.IMREAD_UNCHANGED)
            h, w = img.shape[:2]

            outputs = self.dnn.detect_objects(img)
            detections = self.dnn.calculate_box_dimensions_yolov3(
                outputs, h, w)

        self.metric_service.SetMetrics({"detected-objects": len(detections), "frame-number": frame_number})

        asyncio.create_task(self.metric_service.PushMetrics())

        logging.debug("Made detections: %s", detections)

        await self.call_tracker(tracking.Frame(image, timestamp, frame_number), detections)

        return detection.Empty()

    async def call_tracker(self, frame: tracking.Frame, detections: list):
        tmp = list()
        for detection in detections:
            det = tracking.DetectionsDetection(
                bbox=detection[0], conf=detection[1], clazz=detection[2])
            tmp.append(det)

        await self.stub.add_detections(frame=frame, detections=tmp)

async def main():
    if not ARGS["evaluation_address"] is None:
        metric_service = helper.MetricService(ARGS["evaluation_address"] + "/detection")
    else:
        metric_service = helper.MetricService()
    
    dnn = NeuralNet(ARGS["model"], ARGS["config"], ARGS["classes"], ARGS["confidencethreshold"], ARGS["nmsthreshold"])
    tracker_address = ARGS["tracker_address"].split(':')
    channel = Channel(host=tracker_address[0], port=tracker_address[1])
    detector = Detector(channel, dnn, metric_service)
        
    server = Server([detector])
    with graceful_exit([server]):
        await server.start(port=8080)
        await server.wait_closed()


if __name__ == "__main__":
    # Setup CLI Arguments
    ap = argparse.ArgumentParser(description="Object Detection configuration")
    ap.add_argument("--config",
                    help="path to model config")
    ap.add_argument("--model", required=True,
                    help="path to pre-trained model")
    ap.add_argument("--classes", required=True,
                    help="path to classes detected by model")
    ap.add_argument("--confidencethreshold", type=float, default=0.5,
                    help="minimum probability to filter weak detections")
    ap.add_argument("--nmsthreshold", type=float, default=0.4,
                    help="minimum probability to filter weak detections")
    ap.add_argument("--tracker-address", type=str, default="localhost:8181",
                    help="address for tracker grpc channel")
    ap.add_argument("--evaluation-address", type=str,
                    help="Address and path of evaluation service")
    ap.add_argument("--log-level", type=str, default="INFO",
                    help="Verbosity of generate application logs")
    ap.add_argument("--profiling", action='store_true',
                    help="enable cProfile and export function profiles")
    ARGS = vars(ap.parse_args())

    logging.basicConfig(level=getattr(logging, ARGS["log_level"].upper()))
    logger = logging.getLogger('hpack.hpack').setLevel(logging.ERROR)
    logger = logging.getLogger('urllib3.connectionpool').setLevel(logging.ERROR)

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

