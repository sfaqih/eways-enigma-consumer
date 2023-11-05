local-build:
	rm go.mod;rm go.sum;rm -rf vendor/;git pull;go mod init gitlab.com/wgroup1/enigmaconsumer;go mod tidy;go mod download;
	go mod vendor;go mod verify;go get -d github.com/AplikasiRentasDigital/eways-enigma-master@development;go mod vendor;
dev-build:
	rm go.mod;rm go.sum;rm -rf vendor/;git pull superenigma-dev;go mod init github.com/AplikasiRentasDigital/eways-enigma-consumer;go mod download;
	go mod vendor;go mod verify;go get -d github.com/AplikasiRentasDigital/eways-enigma-master@superenigma-dev;go mod vendor;go mod tidy;rm enigmaconsumer.tar;
	docker build -t enigmaconsumer:latest .;echo Ew4ys1nd0!@5vr;docker save enigmaconsumer:latest > enigmaconsumer.tar;
	microk8s ctr image import enigmaconsumer.tar
prod-build:
	rm go.mod;rm go.sum;rm -rf vendor/;git pull;go mod init github.com/AplikasiRentasDigital/eways-enigma-consumer;go mod tidy;go mod download;
	go mod vendor;go mod verify;go get -d github.com/AplikasiRentasDigital/eways-enigma-master@superenigma;go mod vendor;rm enigmaconsumer.tar;
	docker build -t enigmaconsumer:superenigma .;echo itdev123$$;docker save enigmaconsumer:superenigma > enigmaconsumer.tar;
	microk8s ctr image import enigmaconsumer.tar