# terraform provider cloudflare extended

based on the `next` branch of [next branch](terraform-provider-cloudflare)
which uses the cloudflare-go v3 client.

I needed the following resources for my own projects, hopefully will be deprecated
by the cf provider getting up to date with all of cloudflares services.

Queue Consumers implementation is copied & pasted directly from the `next` branch.

R2 Event Notifications Vectorize, and Workers Script are implemented on their
own in this repo, acceptance tests are passing for them.

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
