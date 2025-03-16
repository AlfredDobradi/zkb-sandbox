server:
	go build -o dist/server ./server/... && ./dist/server

client:
	go build -o dist/client ./client/... && ./dist/client
