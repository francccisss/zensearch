import { Webpage, Segment } from "./types";

// MSS is number in bytes
function createSegments(
  webpages: Array<Webpage>, // webpages queried from database
  MSS: number,
): Array<ArrayBufferLike> {
  const text_encoder = new TextEncoder();
  const encoded_text = text_encoder.encode(JSON.stringify(webpages));
  const data_length = encoded_text.byteLength;
  let currentIndex = 0;
  let segmentCount = Math.trunc(data_length / MSS) + 1; // + 1 to store the remainder
  let segments: Array<ArrayBufferLike> = [];
  let pointerPosition = MSS;

  for (let i = 0; i < segmentCount; i++) {
    let remainingDataLength = Math.abs(currentIndex - data_length);

    let slicedArray = encoded_text.slice(currentIndex, pointerPosition);

    currentIndex += slicedArray.byteLength;
    // Add to offset MSS to point to the next segment in the array
    // manipulate pointerPosition to adjust to lower values using Math.min()

    // Is current data length enough to fit MSS?
    // if so add from current position + MSS
    // else get remaining of the currentDataLength
    pointerPosition += Math.min(MSS, remainingDataLength);
    const payload = new Uint8Array(slicedArray.length);
    payload.set(slicedArray);
    segments.push(newSegment(i, segmentCount, Buffer.from(payload)));
  }
  return segments;
}

function newSegment(
  sequenceNum: number,
  segmentCount: number,
  payload: Buffer,
): ArrayBufferLike {
  // 4 bytes for sequenceNum 4 bytes for totalSegmentsCount
  const sequenceNumBuffer = convertIntToBuffer(sequenceNum);
  const segmentCountBuffer = convertIntToBuffer(segmentCount);
  const headerBuffer = new ArrayBuffer(8);
  const header = new Uint8Array(headerBuffer);
  header.set(Buffer.concat([sequenceNumBuffer, segmentCountBuffer]));
  return Buffer.concat([header, payload]);
}

function convertIntToBuffer(int: number): Buffer {
  const bytes = Buffer.alloc(4);
  bytes.writeIntLE(int, 0, 4);
  console.log(bytes);
  return bytes;
}

export default { createSegments };
