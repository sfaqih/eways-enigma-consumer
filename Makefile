local-build:
	rm go.mod;rm go.sum;rm -rf vendor/;git pull;go mod init gitlab.com/wgroup1/enigmaconsumer;go mod tidy;go mod download;
	go mod vendor;go mod verify;go get -d gitlab.com/wgroup1/enigma@development;go mod vendor;
dev-build:
	rm go.mod;rm go.sum;rm -rf vendor/;git pull superenigma-dev;go mod init gitlab.com/wgroup1/enigma-consumer;go mod download;
	go mod vendor;go mod verify;go get -d gitlab.com/wgroup1/enigma@superenigma-dev;go mod vendor;go mod tidy;rm enigmaconsumer.tar;
	docker build -t enigmaconsumer:latest .;echo Ew4ys1nd0!@5vr;docker save enigmaconsumer:latest > enigmaconsumer.tar;
	microk8s ctr image import enigmaconsumer.tar
prod-build:
	rm go.mod;rm go.sum;rm -rf vendor/;git pull;go mod init gitlab.com/wgroup1/enigma-consumer;go mod tidy;go mod download;
	go mod vendor;go mod verify;go get -d gitlab.com/wgroup1/enigma@superenigma;go mod vendor;rm enigmaconsumer.tar;
	docker build -t enigmaconsumer:superenigma .;echo itdev123$$;docker save enigmaconsumer:superenigma > enigmaconsumer.tar;
	microk8s ctr image import enigmaconsumer.tar