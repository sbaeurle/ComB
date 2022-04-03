FROM python:3.7

RUN mkdir -p /test
WORKDIR /test
ADD . /test/
RUN pip3 install -r requirements.txt

CMD python3 benchmark.py

