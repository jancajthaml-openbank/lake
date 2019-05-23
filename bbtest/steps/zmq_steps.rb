
step "lake recieves :data" do |data|
  ZMQHelper.send(data)
end

step "lake responds with :data" do |data|
  eventually(backoff: 0.5) {
    ok = ZMQHelper.pulled_message?(data)
    expect(ok).to be(true), "message #{data} was not found in #{ZMQHelper.mailbox()}"
  }
  ZMQHelper.ack(data)
end

step "lake performs handshake" do ||
  eventually(backoff: 0.5) {
    ZMQHelper.send("!")
    ok = ZMQHelper.pulled_message?("!")
    expect(ok).to be(true), "handshake was not relayed"
  }
  ZMQHelper.ack("!")
end
