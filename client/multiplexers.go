package client

import (
	"log"

	"github.com/cloudquery/plugin-sdk/schema"
)

func ServiceAccountRegionMultiplexer(table, service string) func(meta schema.ClientMeta) []schema.ClientMeta {
	return func(meta schema.ClientMeta) []schema.ClientMeta {
		var l = []schema.ClientMeta{}
		client := meta.(*Client)

		for account := range client.ClientsAccountRegionMap {
			for region := range client.ClientsAccountRegionMap[account] {
				c := Client{
					Logger:                  client.Logger,
					ClientsAccountRegionMap: client.ClientsAccountRegionMap,
					Account:                 account,
					Region:                  region,
				}

				l = append(l, &c)
			}
		}

		log.Printf("multiplexed clients are: %v", l)

		return l
	}
}
