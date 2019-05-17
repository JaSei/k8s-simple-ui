workflow "New workflow" {
  on = "repository_vulnerability_alert"
  resolves = ["Create an issue"]
}

workflow "New workflow 1" {
  on = "push"
}

action "Create an issue" {
  uses = "JasonEtco/create-an-issue@4ec015aad67f1e9c2f8b6658e1628a2d703b85cb"
}
