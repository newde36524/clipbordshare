.PHONY: client
install:
	go build -o /bin/clipboardshare && systemctl enable clipboardshare-client.service && systemctl restart clipboardshare-client.service && systemctl status clipboardshare-client.service

.PHONY: server
install:
	go build -o /bin/clipboardshare && systemctl enable clipboardshare-server.service && systemctl restart clipboardshare-server.service && systemctl status clipboardshare-server.service
