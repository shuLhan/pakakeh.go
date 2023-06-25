##
## Run autobahn client to test our server.
##
.PHONY: run-autobahn-client
run-autobahn-client:
	sleep 2
	podman run -it --rm \
		--volume="${PWD}/server:/config" \
		--volume="${PWD}/server/testdata:/reports" \
		--name=fuzzingclient \
		docker.io/crossbario/autobahn-testsuite \
		wstest --mode fuzzingclient --spec=/config/fuzzingclient.json

.PHONY: run-test-server
run-test-server:
	go run ./server &

.PHONY: test-server
test-server: run-test-server run-autobahn-client
	go run ./server shutdown
	go run ./server reports

##
## Run autobahn server to test our client in the background.
##
.PHONY: run-autobahn-server
run-autobahn-server:
	-podman stop fuzzingserver
	podman run -it --rm --detach \
		--volume "${PWD}/client:/config" \
		--volume "${PWD}/client/testdata:/reports" \
		--publish 9001:9001 \
		--publish 8080:8080 \
		--name fuzzingserver \
		docker.io/crossbario/autobahn-testsuite \
		wstest --mode fuzzingserver --spec=/config/fuzzingserver.json

.PHONY: test-client
test-client: run-autobahn-server
	sleep 2
	go run ./client
	podman stop fuzzingserver
