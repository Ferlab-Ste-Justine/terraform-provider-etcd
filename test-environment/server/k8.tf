resource "kubernetes_config_map" "etcd_scripts" {
  metadata {
    name = "etcd-scripts"
  }

  data = {
    "entrypoint.sh" = "${file("${path.module}/run-etcd.sh")}"
    "bootstrap-auth.sh" = "${file("${path.module}/bootstrap-etcd-auth.sh")}"
  }
}

resource "kubernetes_secret" "certificates" {
  metadata {
    name = "etcd-certificates"
  }

  data = {
    "ca.pem" = module.ca.certificate
    "server.key" = tls_private_key.key.private_key_pem
    "server.pem" = tls_locally_signed_cert.certificate.cert_pem
    "root.key" = module.root_certificate.key
    "root.pem" = module.root_certificate.certificate
  }
}

resource "kubernetes_deployment" "etcd" {
  metadata {
    name = "etcd"
    labels = {
      app = "etcd"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "etcd"
      }
    }

    template {
      metadata {
        labels = {
          app = "etcd"
        }
      }

      spec {
        volume {
          name = "scripts"
          config_map {
            name = "etcd-scripts"
            default_mode = "0555"
          }
        }

        volume {
          name = "certificates"
          secret {
            secret_name = "etcd-certificates"
            default_mode = "0400"
          }
        }

        container {
          image = "quay.io/coreos/etcd:v3.4.15"
          name  = "etcd"
          command = ["/opt/scripts/entrypoint.sh"]

          volume_mount {
            mount_path = "/opt/scripts"
            read_only = true
            name = "scripts"
          }

          volume_mount {
            mount_path = "/opt/certs"
            read_only = true
            name = "certificates"
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_secret.certificates,
    kubernetes_config_map.etcd_scripts,
  ]
}

resource "kubernetes_service" "etcd" {
  metadata {
    name = "etcd"
  }
  spec {
    selector = {
      app = "etcd"
    }
    port {
      port        = 2379
      target_port = 2379
      node_port = 32379
    }

    type = "NodePort"
  }
}