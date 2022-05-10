package main

import (
	"log"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

func main() {
	c, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{
		Groups: unix.RTMGRP_IPV4_IFADDR,
	})
	if err != nil {
		log.Fatalf("failed to dial netlink: %v", err)
	}
	defer c.Close()

	for {
		msgs, err := c.Receive()
		if err != nil {
			log.Println(err)
			break
		}
		for _, msg := range msgs {
			var res netlink.Message
			if err := (&res).UnmarshalBinary(msg.Data[4:]); err != nil {
				log.Fatalf("failed to unmarshal response: %v", err)
			}
			log.Printf("%+v", res)
		}
	}
}
