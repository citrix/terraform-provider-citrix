provider "citrix" {
  customer_id                 = "" # set your customer id
  client_id                   = ""
  client_secret               = "" # API key client id and secret are needed to interact with Citrix Cloud APIs. These can be created/found under Identity and Access Management > API Access
  environment                 = "Production" # use "Japan" for Citrix Cloud customers in Japan region
}
