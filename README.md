# Cloudflare-DynDNS

Golang implementation of DynDNS on CloudFlare.  Update a dns record to be your specific public IP (which might change over time).

# Usage

  - `cp config.env.example config.env` - Create config.env file
  - `nano config.env` - Edit file with your creds + subdomain + domain
  - `make docker-build` - Compile the go code + build container
  - `make docker-run` - Run docker container with config.env as environment vars in container

# Todo
  - [ ] Implement glog / logging levels
  - [ ] Clean up main.go
  - [ ] Clean Makefile/scripts
  - [X] Create k8s secret generator from config.env file?
    - [X] Make secret mounted as volume/file, pass --env-file to main.go
  - [ ] Dockerhub auto tag images 
