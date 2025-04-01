# I. BUILD NEW LIBRARY FILES: 
### 1. install go and make sure go1.22

```bash
export GO_HOME=/usr/local/1.22
export PATH=$GO_HOME/bin:$PATH

go version
go version go1.1.22 linux/amd64
```
### 2. install dependence and build binary file:
```bash
# install dependence
cd kibanaler/src
go mod tidy

# build binary file
cd kibanaler/src
go build -o ../bin/run .
```

# II. GETTING STARTED
You only need the `./bin/run` and `./bin/.env`.  Here's how:

### 1. create your own .env
```bash
cd ./bin
cp env.sample .env
```

### 2. fill your own .env
Fill out the details of `.env`.

### 3. run the app
```bash
Type `./run` to start application
```

# III. Config influxdb:

### 1. create `config` and `access`:
```bash
# create config
influx config create \
  --config-name duongdx-influx \
  --host-url http://localhost:8086 \
  --org duongdx \
  --token my-secret-token \
  --active

# list config:
influx config list    
  
Active	Name		URL			Org
*	admin-influx	http://localhost:8086	duongdx
	duongdx-influx	http://localhost:8086	duongdx

# active config:
influx config set --active --config-name duongdx-influx
influx config set -a -c duongdx-influx
```

### 2. create `monitor` bucket:
```bash
influx bucket create --name monitoring --retention 30d --org duongdx --token my-secret-token
```

### 3. interact with influxdb: 
```bash
# list bucket command
influx bucket list

# result
ID			Name		Retention	Shard group duration	Organization ID		Schema Type
0a034cf152f1b67f	_monitoring	168h0m0s	24h0m0s			c84e1a8e43c246a0	implicit
03a36eb63fe8d88a	_tasks		72h0m0s		24h0m0s			c84e1a8e43c246a0	implicit
2ee8ca2711ea9123	monitoring	720h0m0s	24h0m0s			c84e1a8e43c246a0	implicit
c80a234ba8828ead	mybucket	infinite	168h0m0s		c84e1a8e43c246a0	implicit
```

### 4. interact with influxdb:
```bash
# get current date:
CURRENT_DATE=$(date +%s)

# insert data
influx write \
  --o duongdx \
  --bucket monitoring \
  --precision s \
  "cpu_usage,host=server01,region=ap-southeast-1 value=80.2 $CURRENT_DATE"
```

### 5. query with influxdb:
```bash
influx query '
from(bucket: "monitoring")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "cpu_usage")' \
  --org duongdx --token my-secret-token
```

# IV. Deploy to lambda function:

### 1. build artifact:
```bash
# build artifact
GOOS=linux GOARCH=arm64 go build -o main run.go
```

### 2. zip code:
```bash
zip function.zip main
```

### 3. create lambda function:
```bash
aws lambda create-function --function-name myFunction \
  --runtime provided.al2023 \
  --handler bootstrap \
  --architectures arm64 \
  --role arn:aws:iam::111122223333:role/lambda-ex \
  --zip-file fileb://function.zip
```

### 4. invoke function:
```bash
aws lambda invoke --function-name myFunction response.json
cat response.json
```