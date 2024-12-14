# 安装Shifu
## 下载安装Shifu Demo
```bash
wxh@DESKTOP-70QQ1RV:~$ curl -sfL https://raw.githubusercontent.com/Edgenesis/shifu/main/test/scripts/shifu-demo-install.sh | sudo sh -
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  698M  100  698M    0     0  4406k      0  0:02:42  0:02:42 --:--:-- 2092k
shifu_demo_aio_linux_amd64.tar.gz
test/scripts/deviceshifu-demo-aio.sh
running on linux/amd64, tar name should be shifu_demo_aio_linux_amd64.tar
running demo
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   194    0   194    0     0   1985      0 --:--:-- --:--:-- --:--:--  2000
./
./util_dir/
./util_dir/kubectl
./util_dir/kind
./shifu/
./shifu/shifu_install.yml
./shifu/demo_device/

...

edgedevice.shifu.edgenesis.io/edgedevice-agv created
service/agv created
configmap/agv-configmap-0.0.1 created
deployment.apps/deviceshifu-agv-deployment created
service/deviceshifu-agv created
Finished setting up Demo !
```

## 确认Shifu已启动
通过`sudo kubectl get pods -A`查看集群中所有正在运行的 Pod，注意到命名空间`shifu-crd-system`和`<font style="color:rgb(25, 60, 71);">deviceshifu</font>`<font style="color:rgb(25, 60, 71);">下的控制器和数字孪生正在运行，表明Shifu成功安装启动。</font>

```bash
wxh@DESKTOP-70QQ1RV:~$ sudo kubectl get pods -A
NAMESPACE            NAME                                            READY   STATUS    RESTARTS   AGE
devices              agv-58c86865d7-p5ffq                            1/1     Running   0          10m
deviceshifu          deviceshifu-agv-deployment-69c5444df4-4jg65     1/1     Running   0          10m
kube-system          coredns-6f6b679f8f-dh49q                        1/1     Running   0          11m
kube-system          coredns-6f6b679f8f-fdhgk                        1/1     Running   0          11m
kube-system          etcd-kind-control-plane                         1/1     Running   0          11m
kube-system          kindnet-ztcj4                                   1/1     Running   0          11m
kube-system          kube-apiserver-kind-control-plane               1/1     Running   0          11m
kube-system          kube-controller-manager-kind-control-plane      1/1     Running   0          11m
kube-system          kube-proxy-xh5xn                                1/1     Running   0          11m
kube-system          kube-scheduler-kind-control-plane               1/1     Running   0          11m
local-path-storage   local-path-provisioner-57c5987fd4-qx4cl         1/1     Running   0          11m
shifu-crd-system     shifu-crd-controller-manager-79d4f569d9-tkbxw   2/2     Running   0          10m
```



# 运行酶标仪的数字孪生
## 准备工作
启动nginx与deviceShifu交互

```bash
wxh@DESKTOP-70QQ1RV:~$ cd shifudemos/
wxh@DESKTOP-70QQ1RV:~/shifudemos$ sudo kubectl run --image=nginx:1.21 nginx
kubectl get pods -A | grep nginx[sudo] password for wxh:
Sorry, try again.
[sudo] password for wxh:
pod/nginx created
wxh@DESKTOP-70QQ1RV:~/shifudemos$ sudo kubectl get pods -A | grep nginx
default              nginx                                           1/1     Running   0          23s
```



## 创建并启动酶标仪的数字孪生
```bash
wxh@DESKTOP-70QQ1RV:~/shifudemos$ sudo kubectl apply -f run_dir/shifu/demo_device/edgedevice-plate-reader
configmap/plate-reader-configmap-0.0.1 created
deployment.apps/deviceshifu-plate-reader-deployment created
service/deviceshifu-plate-reader created
deployment.apps/plate-reader created
edgedevice.shifu.edgenesis.io/edgedevice-plate-reader created
service/plate-reader created
```

查看酶标仪启动情况：

```bash
wxh@DESKTOP-70QQ1RV:~/shifudemos$ sudo kubectl get pods -A | grep plate
devices              plate-reader-6d58f85f9c-q8s5v                          1/1     Running   0          114s
deviceshifu          deviceshifu-plate-reader-deployment-67b594886d-zkcqc   1/1     Running   0          114s
```

`sudo kubectl exec -it nginx -- bash`进入nginx，通过nginx与酶标仪的数字孪生交互，获得测量结果：

```bash
root@nginx:/# curl "deviceshifu-plate-reader.deviceshifu.svc.cluster.local/get_measurement"
1.99 0.47 1.55 2.46 0.70 2.68 1.08 1.18 1.15 2.34 2.55 2.37
1.13 1.17 2.88 1.28 2.02 0.92 0.54 1.11 2.80 1.51 2.04 1.88
0.32 2.16 1.88 1.91 2.51 1.82 2.43 1.43 0.50 1.88 1.21 0.15
0.76 0.19 0.53 2.38 0.14 2.29 1.99 1.81 0.39 2.94 1.17 2.68
1.58 0.14 2.91 0.41 2.02 1.14 1.53 1.07 2.08 1.98 1.22 2.10
2.48 1.57 0.32 1.91 0.49 2.93 2.26 2.15 1.66 0.32 1.00 2.71
2.90 1.48 0.02 1.65 2.03 0.45 0.26 0.48 1.23 1.09 2.32 2.78
2.09 1.34 0.32 0.68 1.18 0.34 2.31 0.55 1.22 2.95 0.83 2.42
```

# 编写Go应用
编写go应用，定时轮询`http://deviceshifu-plate-reader.deviceshifu.svc.cluster.local/get_measurement`接口返回的矩阵数据，并打印平均值。设置默认轮询间隔为5s，可以通过环境变量传入自定义值。通过for循环和定时器每次计算矩阵数据的平均值。使用`log.Printf`打印日志，方便通过`kubectl logs`查看。

```go
// 默认轮询间隔
const defaultPollInterval = 5 * time.Second

func main() {
    // 获取轮询间隔（通过环境变量配置）
    pollInterval := getPollInterval()

    // 接口 URL
    targetUrl := "http://deviceshifu-plate-reader.deviceshifu.svc.cluster.local/get_measurement"

    for {
        // 获取酶标仪返回的矩阵数据
        matrix, err := fetchMatrix(targetUrl)
        if err != nil {
            log.Printf("Error fetching matrix: %v", err)
            time.Sleep(pollInterval)
            continue
        }

        // 计算平均值
        average := calculateAverage(matrix)
        log.Printf("Average: %.2f", average)

        // 等待下一轮
        time.Sleep(pollInterval)
    }
}

// 获取轮询间隔
func getPollInterval() time.Duration {
    envInterval := os.Getenv("POLL_INTERVAL")
    if envInterval == "" {
        return defaultPollInterval
    }

    interval, err := strconv.Atoi(envInterval)
    if err != nil || interval <= 0 {
        log.Printf("Invalid POLL_INTERVAL: %v, using default: %v", err, defaultPollInterval)
        return defaultPollInterval
    }
    return time.Duration(interval) * time.Second
}

// 从目标 URL 获取矩阵数据
func fetchMatrix(url string) ([][]float64, error) {
    res, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }

    return parseMatrix(string(body))
}

// 解析矩阵数据
func parseMatrix(data string) ([][]float64, error) {
    lines := strings.Split(strings.TrimSpace(data), "\n")
    matrix := make([][]float64, len(lines))

    for i, line := range lines {
        values := strings.Fields(line)
        row := make([]float64, len(values))

        for j, val := range values {
            num, err := strconv.ParseFloat(val, 64)
            if err != nil {
                return nil, err
            }
            row[j] = num
        }
        matrix[i] = row
    }
    return matrix, nil
}

// 计算矩阵的平均值
func calculateAverage(matrix [][]float64) float64 {
    var sum float64
    var count int

    for _, row := range matrix {
        for _, val := range row {
            sum += val
            count++
        }
    }

    if count == 0 {
        return 0
    }
    return sum / float64(count)
}

```



编写Dockerfile

```dockerfile
# 使用官方 Golang 镜像
FROM golang:1.22

# 设置工作目录
WORKDIR /app

# 将代码复制到容器中
COPY . .

# 编译应用程序
RUN go build -o app .

# 设置运行时环境变量（可在部署时覆盖）
ENV POLL_INTERVAL=5

# 容器启动命令
CMD ["./app"]

```

创建镜像 `docker build -t plate-poller:1.0 -f Dockerfile .`

![](https://cdn.nlark.com/yuque/0/2024/png/43482173/1734152127920-a2c22f06-df7b-490b-a3e8-731e6442c7b7.png)



修改镜像名符合dockerhub远程仓库的规范，push到仓库

![](https://cdn.nlark.com/yuque/0/2024/png/43482173/1734152340089-639897e6-ded9-487b-b8f9-8c9ac0a9cf20.png)

创建pod运行

![](https://cdn.nlark.com/yuque/0/2024/png/43482173/1734152635267-665bba68-10a2-4434-9ad5-7bdb769bcc39.png)

通过`kubectl logs`查看日志，可以看到每5s成功打印出平均值

![](https://cdn.nlark.com/yuque/0/2024/png/43482173/1734148345416-82b57a25-8d09-4c7f-9f0b-5756d3672e34.png)

# 总结
Shifu的文档非常详尽，数字孪生的交互以及创建与 deviceShifu 交互的应用都有示例参考。在笔试过程中没有遇到太多问题，主要还是关于k8s的操作不够熟练，通过文档以及`kubectl describe pod`命令进行调试，最终成功运行。







