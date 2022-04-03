# ComB: A Flexible, Application-Oriented Benchmark for Edge Computing

Edge computing is an attractive platform where applications, previously hosted in the cloud, shift parts of their workload on resources
closer to the users. The field is still in its nascent stages with significant ongoing innovation in small form-factor hardware designed to
operate at the edge. However, the increased hardware heterogeneity
at the edge makes it difficult for application developers to determine if their workloads will operate as desired. Simultaneously,
edge providers have to make expensive deployment choices for
the “correct” hardware that will remain suitable for the near future.
We present ComB, an application-oriented benchmarking suite for
edge that assists early adopters in evaluating the suitability of an
edge deployment. ComB is flexible, extensible, and incorporates a
microservice-based video analytics pipeline as default workload to
measure underlying hardware’s compute and networking capabilities accurately. Our evaluation on a heterogeneous testbed shows
that ComB enables both providers and developers to understand
better the runtime capabilities of different hardware configurations
for supporting operations of applications designed for the edge.

---

If you use ComB for your own research, please cite our [paper](https://doi.org/10.1145/3517206.3526269).

## Requirements

- GO 1.16+
- Python 3.8+

## Getting Started

Build Metric Collection System:

```
cd metrics
go build
```

Copy [TrackEval](https://github.com/JonathonLuiten/TrackEval) code into `metrics/modules/TrackEval` (or clone the repository).

Build Orchestration:

```
cd orchestration
go build
```

Run the benchmark:

```
./metrics --config 
./orchestration --config
```

Run both executable with the `--config` flag pointing to the configuration corresponding to your workload.

## Metric Configuration

```
RootFolder: results # Output Folder
DateFormat: 20060102_150405 # Date Format to structure separate benchmark runs (for layout see https://pkg.go.dev/time#pkg-constants)
PlottingScript: plotting.py # Script used to plot (currently under redevelopment)
Endpoints:
  - Name: Name
    Url: /route # HTTP Route for the benchmark endpoint
    Module: ModuleName # Evaluation Module used
    Config: # Additional configuration for the module
    Fields: ["field"] # List of JSON fields received from the workload
    Outputs: ["output.(txt/csv)"] # Name of the output file
    Metrics: # List of Metrics and their (possible) aggregations
        - Metric: [Aggregations]
```

## Benchmark Configuration

```
Evaluation: http://127.0.0.1:8000 # Base Address of the metric collection service
SSH:
  User: ubuntu # SSH user
  KeyFile: keypath # SSH keyfile
  Commands:
    - cpu: "docker run --name {{.Name}}{{range $i, $v := .Ports}} -p {{$v}}{{end}}{{range $i, $m := .Mounts}} -v {{$m}}{{end}} {{.Image}}:{{.Tag}} {{.Command}}" # technology specific command to start docker container
NodeGroups:
  - Name: APU
    Arch: x86 # Supported Architecture of NodeGroup
    Capabilities: [cpu, opencl] # Supported technologies of NodeGroup
    Nodes: [127.0.0.1] # List of Node addresses
    NodeCapacity: 2 # Maximum workload capacity per node
Workload:
  - Name: Tracking
    Image: sbaeurle/tracking # docker image name
    Ports:
      - 8181:8181 # exposed ports
    Tags: [cpu] # available tags/technologies
    Arch: [x86, arm64] # available architectures
    Command: python3 object_tracking.py --tracker kcf --log-level DEBUG --evaluation-address {{.Evaluation}} # Command to startup the image (fields with {{}} will be replaced using GOs text/template)
```

## Results

Results are written in the configured `RootFolder` in form of `subfolder/` (based on the time layout) and `run00n` based on the current number of runs.

## Workload

- `benchmarks/tracking_pipeline`: This includes the code for the default video analytics pipeline.
- `benchmarks/video_source`: This contains the adapted code for the RTSP server to serve H264 videos as IP stream.
- `benchmarks/network_overhead`: A sample workload we included from another, internal evaluation project. Highlights the ease of adjusting our benchmark suite.