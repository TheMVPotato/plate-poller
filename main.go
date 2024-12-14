package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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
