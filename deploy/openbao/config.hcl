storage "raft" {
  path    = "/openbao/file"
  node_id = "node1"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = true
}

cluster_addr      = "http://127.0.0.1:8201"
api_addr          = "http://0.0.0.0:8200"
default_lease_ttl = "168h"
max_lease_ttl     = "720h"
ui = true