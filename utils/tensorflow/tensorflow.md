
# Appendix.1 Tensorflow for Go installation guide
## 1.Install Tensorflow C lib

- Linux (For CPU)	
	 
> https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz

- Linux (For GPU)	

> https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz


```bash
wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz

sudo tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz
sudo ldconfig
```
## 2.Verify tensorflow C lib

```c
#include <stdio.h>
#include <tensorflow/c/c_api.h>

int main() {
  printf("Hello from TensorFlow C library version %s\n", TF_Version());
  return 0;
}
```

```bash
gcc hello_tf.c -ltensorflow -o hello_tf
./hello_tf
```

## 3. Install protobuf

```bash
git clone https://github.com/protocolbuffers/protobuf.git
cd protobuf/ 
./autogen.sh 
./configure
make 
sudo make install
sudo ldconfig 
protoc -h 

go get -v -u github.com/golang/protobuf/proto
go get  -v -u github.com/golang/protobuf/protoc-gen-go
```

## 3. Install Tensorflow for Go

```bash
go get -d github.com/tensorflow/tensorflow/tensorflow/go
go generate github.com/tensorflow/tensorflow/tensorflow/go/op

export LIBRARY_PATH=$LIBRARY_PATH:/usr/local/lib
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:/usr/local/lib
```

## 4. Verify

```bash
go test github.com/tensorflow/tensorflow/tensorflow/go
```

```go
package main

import (
    tf "github.com/tensorflow/tensorflow/tensorflow/go"
    "github.com/tensorflow/tensorflow/tensorflow/go/op"
    "fmt"
)

func main() {
    // Construct a graph with an operation that produces a string constant.
    s := op.NewScope()
    c := op.Const(s, "Hello from TensorFlow version " + tf.Version())
    graph, err := s.Finalize()
    if err != nil {
        panic(err)
    }

    // Execute the graph in a session.
    sess, err := tf.NewSession(graph, nil)
    if err != nil {
        panic(err)
    }
    output, err := sess.Run(nil, []tf.Output{c}, nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(output[0].Value())
}

```

```bash
go build main.go
./main

2021-07-21 08:42:45.919364: I tensorflow/core/platform/cpu_feature_guard.cc:142] This TensorFlow binary is optimized with oneAPI Deep Neural Network Library (oneDNN) to use the following CPU instructions in performance-critical operations:  AVX2 FMA
To enable them in other operations, rebuild TensorFlow with the appropriate compiler flags.
2021-07-21 08:42:45.924877: I tensorflow/core/platform/profile_utils/cpu_utils.cc:112] CPU Frequency: 3100005000 Hz
Hello from TensorFlow version 2.4.0

```




