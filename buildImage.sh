mvn clean install
swagger generate server -f dsb-swagger.yaml -A mysqldsb
docker build -t ocopea/mysql-k8s-dsb -f Dockerfile ../
