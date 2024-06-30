# Graduation Thesis Backend

Project Golang này là phần backend của ĐATN với đề tài "Ứng dụng nhắn tin mã hóa đầu cuối"

## Yêu cầu hệ thống

- Golang 1.21 hoặc mới hơn
- Git

## Cài đặt

### Bước 1: Cài đặt Golang

Nếu bạn chưa cài đặt Golang, vui lòng tải và cài đặt từ trang web chính thức: [https://golang.org/dl/](https://golang.org/dl/)

### Bước 2: Kiểm tra cài đặt Golang

Sau khi cài đặt, mở terminal (hoặc Command Prompt trên Windows) và kiểm tra phiên bản Golang:

```sh
go version

### Bước 3: Clone Repository

```sh
git clone https://github.com/ngnangcuong/graduation-thesis-backend.git
cd graduation-thesis-backend

### Bước 4: Cài đặt các dependencies

Sử dụng câu lệnh trên terminal (hoặc Command Prompt trên Windows) để cài đặt dependencies:
```sh
go mod download
go mod tidy

### Bước 5: Build các image của các Service

Trong thư mục deployments có chứa các file Dockerfile để build các image ứng với mỗi Service trong hệ thống. Sử dụng câu lệnh để build các image, ví dụ với User Service:

```sh
docker build -f deployments/Dockerfile.User -t user_service:1.0 .

### Bước 6: Chạy các image bằng Docker compose

Trong thư mục deployments có chứa các file docker-compose.yml để chạy các image ứng với mỗi Service trong hệ thống. Sử dụng câu lệnh để chạy các Service, ví dụ với User Service và Group Service:

```sh
docker compose -f deployments/user-group.yml up -d
