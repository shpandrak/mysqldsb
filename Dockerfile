# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/ocopea
ADD mysqldsb/target/mysql/mysqldump /go/bin/mysqldump
ADD mysqldsb/target/mysql/mysql /go/bin/mysql

# Build command inside the container.
# (You may fetch or manage dependencies here,
RUN go install ocopea/mysqldsb/cmd/mysqldsb-server

ENTRYPOINT /go/bin/mysqldsb-server

EXPOSE 8000
