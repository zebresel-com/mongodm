IMAGE = mongodm
APP_DIR = /go/src/mongodm
RUN = docker run -it --rm -v $(PWD):$(APP_DIR) $(IMAGE)
ARGS = $(filter-out $@,$(MAKECMDGOALS))

test:
	docker-compose run test

dep:
	echo $(ARGS)
	$(RUN) dep $(ARGS)

%:
	@:

.PHONY:
	test dep
