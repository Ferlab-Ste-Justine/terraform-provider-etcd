resource "tls_private_key" "ca" {
  algorithm   = "RSA"
  rsa_bits    = 4096
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm   = tls_private_key.ca.algorithm
  private_key_pem = tls_private_key.ca.private_key_pem

  subject {
    common_name  = var.common_name
  }

  validity_period_hours = 100*365*24
  early_renewal_hours = 99*365*24

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "cert_signing",
  ]

  is_ca_certificate = true
}