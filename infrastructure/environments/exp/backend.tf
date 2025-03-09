terraform {
  backend "gcs" {
    bucket = "coda-exp-tfstate"
    prefix = "terraform/state/exp"
  }
}
