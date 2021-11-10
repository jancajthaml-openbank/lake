use tokio::io::AsyncReadExt;
use tokio::net::tcp::ReadHalf;

// https://v0-1--tokio.netlify.app/docs/io/reading_writing_data/

pub async fn read_frame<'a>(reader: &mut ReadHalf<'a>) -> Result<(Vec<u8>, bool, bool), String> {
    let frame_type = match reader.read_u8().await {
        Ok(v) => v,
        Err(_) => return Err("Failed to read first byte of Frame".to_owned()),
    };

    let mut frame_buffer = match frame_type {
        0x00 | 0x01 | 0x04 => match reader.read_u8().await {
            Ok(size) => vec![0 as u8; size as usize],
            Err(_) => return Err("Failed to read size of Frame".to_owned()),
        },
        0x02 | 0x03 | 0x06 => {
            match reader.read_u64().await {
                Ok(size) => {
                    if size >= usize::MAX as u64 {
                        return Err("Failed to read size of Frame, size is larger than this system can handle".to_owned());
                    } else {
                        vec![0 as u8; size as usize]
                    }
                }
                Err(_) => return Err("Failed to read size of Frame".to_owned()),
            }
        }
        _ => return Err("Invalid frame header size".to_owned()),
    };

    let is_last = match frame_type {
        0x00 | 0x02 | 0x04 | 0x06 => true,
        _ => false,
    };

    let is_cmd = match frame_type {
        0x04 | 0x06 => true,
        _ => false,
    };

    let n = match reader.read_exact(&mut frame_buffer).await {
        Ok(v) => v,
        Err(_) => return Err("Error Reading TCP Packet".to_owned()),
    };

    if n != frame_buffer.len() {
        return Err("Short read".to_owned());
    }

    //println!("red total {} bytes", 1 + 1 + n);

    Ok((frame_buffer, is_last, is_cmd))
}

// 13000 / s
pub async fn read_message<'a>(reader: &mut ReadHalf<'a>) -> Result<Vec<u8>, String> {
    let mut frame_buffer = vec![0 as u8; 69];

    match reader.read_exact(&mut frame_buffer).await {
        Ok(_) => Ok(frame_buffer),
        Err(_) => Err("Error Reading TCP Packet".to_owned()),
    }
}

// 9000 / s
pub async fn read_message_actual<'a>(reader: &mut ReadHalf<'a>) -> Result<Vec<u8>, String> {
    let mut frame = match read_frame(reader).await {
        Ok((_, _, true)) => return Err("Command red instead of Message".to_owned()),
        Ok((a, true, false)) => return Ok(a),
        Ok((a, false, false)) => a,
        Err(_) => return Err("Failed to read Frame".to_owned()),
    };
    loop {
        let (next_frame, is_last) = match read_frame(reader).await {
            Ok((_, _, true)) => {
                return Err("Next Frame of Message changed its type to Command".to_owned())
            }
            Ok((a, b, _)) => (a, b),
            Err(_) => return Err("Failed to read Frame".to_owned()),
        };
        frame.extend_from_slice(&next_frame);
        if is_last {
            return Ok(frame);
        }
    }
}
