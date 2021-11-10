use tokio::io::AsyncReadExt;
use tokio::io::AsyncWriteExt;
use tokio::net::tcp::ReadHalf;
use tokio::net::tcp::WriteHalf;
use tokio::net::TcpListener;
use tokio::sync::mpsc;
use tokio::sync::mpsc::UnboundedReceiver;
use tokio::task::JoinHandle;

use crate::codec::{read_frame, read_message};

pub struct PullRoutine {
    receiver: UnboundedReceiver<Vec<u8>>,
    handler: JoinHandle<u8>,
}

impl Drop for PullRoutine {
    fn drop(&mut self) {
        drop(&self.handler)
    }
}

// FIXME https://en.wikipedia.org/wiki/Fair_queuing

impl PullRoutine {
    #[must_use]
    pub fn new(port: u16) -> PullRoutine {
        let (sender, receiver) = mpsc::unbounded_channel::<Vec<u8>>();

        let handler = tokio::spawn(async move {
            let pull_listener = TcpListener::bind(("127.0.0.1", port))
                .await
                .expect("Failed to bind RAW TCP PULL socket");

            println!("Pull Loop Started");
            loop {
                let incoming = pull_listener.accept().await;
                println!("Connection Accepted");
                let s2 = sender.clone();
                tokio::spawn(async move {
                    let mut raw_socket = incoming
                        .map(|(raw_socket, _remote_addr)| raw_socket)
                        .and_then(|raw_socket| raw_socket.set_nodelay(true).map(|_| raw_socket))
                        .and_then(|raw_socket| raw_socket.set_linger(None).map(|_| raw_socket))
                        .unwrap();

                    let (mut reader, mut writer) = raw_socket.split();

                    greetings(&mut reader, &mut writer)
                        .await
                        .expect("Failed to Greet");
                    handshake(&mut reader, &mut writer)
                        .await
                        .expect("Failed to Handshake");

                    println!("Reading from TCP connection loop");

                    let mut detect_disconnect = [0 as u8; 1];

                    loop {
                        match read_message(&mut reader).await {
                            Ok(message) => match s2.send(message) {
                                Ok(_) => {}
                                Err(err) => {
                                    println!("Unable to add message to inner queue {:?}", err)
                                }
                            },
                            Err(err) => match reader.peek(&mut detect_disconnect).await {
                                Ok(0) => {
                                    println!("Connection closed!");
                                    break;
                                }
                                Ok(_) => {
                                    println!("read_frame ERR {}", err);
                                }
                                Err(err) => {
                                    println!("read_frame ERR {}", err);
                                    println!("Connection closed? {}", err);
                                    break;
                                }
                            },
                        };
                    }
                });
            }
            println!("Pull Loop Stopped"); // FIXME unreachalbe no coordinated shutdown here
        });

        PullRoutine {
            receiver: receiver,
            handler: handler,
        }
    }

    pub async fn recv(&mut self) -> Option<Vec<u8>> {
        self.receiver.recv().await
    }
}

async fn greetings<'a>(
    reader: &mut ReadHalf<'a>,
    writer: &mut WriteHalf<'a>,
) -> Result<(), String> {
    let mut greetings_signature_buffer = [0 as u8; 10];

    match reader.read_exact(&mut greetings_signature_buffer).await {
        Ok(10) => {}
        Ok(_) => return Err("Short read".to_owned()),
        Err(_) => return Err("Error Reading TCP Packet".to_owned()),
    };

    if greetings_signature_buffer[0] != 0xff || greetings_signature_buffer[9] != 0x7f {
        return Err("Wrong Signature".to_owned());
    }

    let mut greetings_reply = [0 as u8; 64];
    greetings_reply[0] = 0xff;
    greetings_reply[9] = 0x7f;
    greetings_reply[10] = 0x03;
    greetings_reply[12] = b'N';
    greetings_reply[13] = b'U';
    greetings_reply[14] = b'L';
    greetings_reply[15] = b'L';

    match writer.write_all(&greetings_reply).await {
        Ok(_) => {}
        Err(_) => return Err("Failed to Send Greeting to Client".to_owned()),
    }

    let mut greetings_remaining_buffer = [0 as u8; 54];

    match reader.read_exact(&mut greetings_remaining_buffer).await {
        Ok(54) => {}
        Ok(_) => return Err("Short read".to_owned()),
        Err(_) => return Err("Error Reading TCP Packet".to_owned()),
    };

    let version_major = greetings_remaining_buffer[0];
    if version_major != 3 {
        return Err("Incompatible ZMQ version".to_owned());
    }

    let as_server = greetings_remaining_buffer[33] == 1;
    if as_server {
        return Err("Peer is not a Client".to_owned());
    }

    Ok(())
}

async fn handshake<'a>(
    reader: &mut ReadHalf<'a>,
    writer: &mut WriteHalf<'a>,
) -> Result<(), String> {
    let (command_buffer, _is_last) = match read_frame(reader).await {
        Ok((a, b, true)) => (a, b),
        Ok((_, _, false)) => return Err("Expected Command frame".to_owned()),
        Err(err) => return Err(err),
    };

    let (name_len, command_buffer) = (command_buffer[0] as usize, &command_buffer[1..]);
    let (cmd_name, command_buffer) = (&command_buffer[..name_len], &command_buffer[name_len..]);

    // INFO NULL strategy
    if cmd_name[0] != b'R'
        || cmd_name[1] != b'E'
        || cmd_name[2] != b'A'
        || cmd_name[3] != b'D'
        || cmd_name[4] != b'Y'
    {
        return Err("Expected READY command".to_owned());
    }

    let mut handshake_reply = [0 as u8; 28];
    handshake_reply[0] = 0x04;
    handshake_reply[1] = 26;
    handshake_reply[2] = 0x05;
    handshake_reply[3] = b'R';
    handshake_reply[4] = b'E';
    handshake_reply[5] = b'A';
    handshake_reply[6] = b'D';
    handshake_reply[7] = b'Y';
    handshake_reply[8] = 11;
    handshake_reply[9] = b'S';
    handshake_reply[10] = b'o';
    handshake_reply[11] = b'c';
    handshake_reply[12] = b'k';
    handshake_reply[13] = b'e';
    handshake_reply[14] = b't';
    handshake_reply[15] = b'-';
    handshake_reply[16] = b'T';
    handshake_reply[17] = b'y';
    handshake_reply[18] = b'p';
    handshake_reply[19] = b'e';
    handshake_reply[20] = 0x00;
    handshake_reply[21] = 0x00;
    handshake_reply[22] = 0x00;
    handshake_reply[23] = 0x04;
    handshake_reply[24] = b'P';
    handshake_reply[25] = b'U';
    handshake_reply[26] = b'L';
    handshake_reply[27] = b'L';

    match writer.write_all(&handshake_reply).await {
        Ok(_) => Ok(()),
        Err(_) => Err("Failed to Send Handshake to Client".to_owned()),
    }
}
