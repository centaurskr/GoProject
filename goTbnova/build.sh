echo "goTbnova 종료"
sudo systemctl stop gotbnova.service
echo "process.toml Process 정리"
~/bin/goClsTbnova -f ~/Project/env/process.toml
go build -o goTbnova && strip goTbnova && mv goTbnova ~/bin && sudo systemctl start gotbnova.service
echo "goTbnova 시작"

