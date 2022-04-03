import time
import requests
import sys
import argparse
import json

TEST_REPEAT = 10
NUM_OF_PACKETS = 100
ARGS = dict()

def run_benchmark():
    req_string = "http://"+ARGS["address"]+"/"
    print(req_string)
    for j in range (TEST_REPEAT):	
        tot = 0.0
        print("Attempt: ", j)

        for i in range(NUM_OF_PACKETS):

            start = time.time()
            r = requests.get(req_string)
            content = str(sys.getsizeof(r.text))
            end = time.time()

            tot+=(end-start)
            tot=round(tot, 4)

            if r.status_code != 200:
                print("invalid request")

            data_point = {
                "test_run": j,
                "packet_num": i,
                "latency": tot
            }

            data = json.dumps(data_point)
            requests.post(ARGS["evaluation_address"], data)

        time.sleep(1)


ap = argparse.ArgumentParser()
ap.add_argument("--address", type=str, default="127.0.0.1",
                help="nginx server address")
ap.add_argument("--evaluation-address", type=str,
                help="url of evaluation service. enables periodic metrics.")
ARGS = vars(ap.parse_args())

if __name__ == "__main__":
    run_benchmark()

