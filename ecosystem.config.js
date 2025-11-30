module.exports = {
  apps : [{
    name   : "webhook-host",
    script : "./webhook-server",
    env: {
      PORT: "50800"
    }
  }]
}
