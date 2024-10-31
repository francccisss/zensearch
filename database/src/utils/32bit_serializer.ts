const BYTES_OFFSET = 4;

function serialize(data: any): Buffer {
  const stringified = JSON.stringify({ data });
  const text_encoder = new TextEncoder();
  const encoded_text = text_encoder.encode(stringified);

  //const array_buffer = new ArrayBuffer(encoded_text.length * BYTES_OFFSET);
  //const buffer_view = new Uint32Array(array_buffer);
  //for (let i = 0; i < encoded_text.length; i++) {
  //  buffer_view[i * BYTES_OFFSET] = encoded_text[i];
  //  // Append each element of the encoded_text with an offset of 4 bytes to the 32bit array buffer
  //  // 1st iteration 0 * 4 = buffer_view[0]
  //  // 2nd iteration 1 * 4 = buffer_view[4]
  //  // 3rd iteration 2 * 4 = buffer_view[8]
  //}
  //console.log(buffer_view);

  return encoded_text.buffer as Buffer;
}

//function deserialize<T>(data: Uint32Array): T {
//  // go through each offset from the 32bit array
//
//  return "Hello";
//}

export default { serialize };
