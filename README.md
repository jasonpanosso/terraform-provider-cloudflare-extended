# terraform provider cloudflare extended

based on the `next` branch of [terraform-provider-cloudflare](https://github.com/cloudflare/terraform-provider-cloudflare/tree/next)
which uses the cloudflare-go v3 client.

I needed the following resources for my own projects, hopefully will be deprecated
by the cf provider getting up to date with all of cloudflares services.

Queue Consumers implementation is copied & pasted directly from the `next` branch.

# Additions

- Queue Consumer
  - data source
  - resource
- R2 Event Notification
  - resource
- Workers with all bindings as of 11/05/2024
  - resource
- Vectorize
  - resource
