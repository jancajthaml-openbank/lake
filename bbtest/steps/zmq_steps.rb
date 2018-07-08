
step "lake recieves :data" do |data|
  send_remote_message(data)
end

step "lake responds with :data" do |data|
  with_deadline(timeout: 10) {
    eventually(timeout: 3) {
      expect(remote_mailbox()).to include(data)
      ack_message(data)
    }
  }
end

step "no other messages were recieved" do ||
  expect(remote_mailbox()).to be_empty
end
