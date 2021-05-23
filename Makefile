build-django:
	docker build -t idealism/django-i18n django

run-django:
	docker run -ti -v ${PWD}/django:/idealism-xxm/django -p 8000:8000 --name django-i18n idealism/django-i18n

gen-django-grpc:
	python3 -m grpc_tools.protoc -I django/grpc/proto --python_out=django/grpc/gen --grpc_python_out=django/grpc/gen django/grpc/proto/*.proto

build-gin:
	docker build -t idealism/gin-i18n gin

run-gin:
	docker run -ti -p 8080:8080 --name gin-i18n idealism/gin-i18n

gen-gin-grpc:
	protoc --go_out=plugins=grpc:. *.proto
