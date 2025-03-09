variable "project_id" {
  description = "The project to deploy resources"
  type        = string
}
variable "service_name" {
  description = "The name of the Cloud Run service"
  type        = string
}
variable "region" {
  description = "The region to deploy resources"
  type        = string
}
variable "image" {
  description = "The Docker image to deploy"
  type        = string
}
variable "service_account" {
  description = "The service account to use"
  type        = string
}