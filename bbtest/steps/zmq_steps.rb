
step "lake recieves :data" do |data|
  send_remote_message(data)
end

step "lake responds with :data" do |data|
  eventually(timeout: 10) {
    expect(remote_mailbox()).to include(data)
    ack_message(data)
  }
end

step "no other messages were recieved" do ||
  expect(remote_mailbox()).to be_empty
end
