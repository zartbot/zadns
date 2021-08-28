# zaDNS
zaDNS is a simple DNS forwarder with AI based security policy.

![](https://raw.githubusercontent.com/zartbot/zadns/master/zadns.gif)

## 1. Support Features

### 1.1 Domain based Routing

you can modify `config/route.cfg` to defined a route table to proccess dns request to different server based on domain

`|` is used for seperate domain and dns-server address
`,` is used for multiple dns-server

```
cisco.com|  8.8.8.8,4.4.4.4
google.com| 8.8.8.8
```
### 1.2 Geo based policy
All `A/AAAA` record will trigger GeoIP lookup, you could define your own logic to block some countries or based on Geo Infomation choose the nearest host.
You could also cache the GeoLocation and compare with future result to determine malicious domain

### 1.3 BGP ASN based policy

use BGP ASN to detect CDN or implement SP based traffic engineering

### 1.4 DGA detection
[Domain generation algorithms (DGA) ](https://en.wikipedia.org/wiki/Domain_generation_algorithm) are algorithms seen in various families of malware that are used to periodically generate a large number of domain names that can be used as rendezvous points with their command and control servers.

We have a pre-trained AI model loaded on zaDNS to block such dns query. Detailed Training process(jupyter-notebook) could be found `@utils/dga`
It's based on a simple LSTM neural network
```python
sess = tf.Session()  
K.set_session(sess) 
max_features = 128
model=Sequential()
model.add(Embedding(max_features, 128,name="inputlayer"))
model.add(LSTM(128))
model.add(Dropout(0.5))
model.add(Dense(128, kernel_initializer='uniform', activation='relu'))
model.add(Dense(nb_classes, kernel_initializer='uniform', activation='softmax',name="outputlayer"))
model.compile(loss='categorical_crossentropy', optimizer='adam', metrics=['accuracy'])
model.summary()
```

> DGA algorithm not available on `MAC` and `windows` platform due to tensorflow cross compile issue

### 1.5 Local Host
just like `/etc/hosts`, you could defined private hostname @`config/hosts.cfg`

## 2. Build

```bash
git clone https://github.com/zartbot/zadns
cd zadns

make 
```
build target `./build/zadns`  

### 2.1 Tensorflow C lib(Linux Only)


```bash
#for cpu 
wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz
#for GPU
wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-gpu-linux-x86_64-1.15.0.tar.gz

sudo tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz
sudo ldconfig
```
Or you can use `LD_LIBRARY_PATH` as alternative solution


# 3. Acknowlegement
Appreciate the following opensource project

- github.com/armon/go-radix 
- github.com/miekg/dns 
- github.com/oschwald/geoip2-golang 
- github.com/tensorflow/tensorflow 


# 4. Future work

- support whois correlation
- DNS Cache for SDWAN policy based Routing
- Integrate DNS Server to Cisco Meraki/Viptela SDWAN and IOS-XE/IOS-XR Routing system
- Add IP reputation filter 