
step "lake recieves :data" do |data|
  ZMQHelper.send(data)
end

step "lake responds with :data" do |data|
  eventually() {
    ok = ZMQHelper.pulled_message?(data)
    expect(ok).to be(true), "message #{data} was not found in #{ZMQHelper.mailbox()}"
  }
  ZMQHelper.ack(data)
end

step "no other messages were recieved" do ||
  expect(ZMQHelper.mailbox()).to be_empty
end
