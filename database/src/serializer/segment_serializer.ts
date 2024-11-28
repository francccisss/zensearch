import { Webpage, Segment } from "../utils/types";

// MSS is number in bytes
function createSegments(
  webpages: Array<Webpage>, // webpages queried from database
  MSS: number,
): Array<Buffer> {
  const textEncoder = new TextEncoder();
  const encodedText = textEncoder.encode(JSON.stringify(webpages));
  const dataLength = encodedText.byteLength;
  let currentIndex = 0;
  let segmentCount = Math.trunc(dataLength / MSS) + 1; // + 1 to store the remainder
  let segments: Array<Buffer> = [];
  let pointerPosition = MSS;

  for (let i = 0; i < segmentCount; i++) {
    let slicedArray = encodedText.slice(currentIndex, pointerPosition);

    currentIndex += slicedArray.byteLength;
    // Add to offset MSS to point to the next segment in the array
    // manipulate pointerPosition to adjust to lower values using Math.min()

    // Is current data length enough to fit MSS?
    // if so add from current position + MSS
    // else get remaining of the currentDataLength
    pointerPosition += Math.min(MSS, Math.abs(currentIndex - dataLength));
    const payload = Buffer.alloc(slicedArray.length);
    payload.set(slicedArray);
    segments.push(newSegment(i, segmentCount, Buffer.from(payload)));
  }
  return segments;
}

function newSegment(
  sequenceNum: number,
  segmentCount: number,
  payload: Buffer,
): Buffer {
  // 4 bytes for sequenceNum 4 bytes for totalSegmentsCount
  const headerBuffer = Buffer.alloc(8);
  const sequenceNumBuffer = convertIntToBuffer(sequenceNum);
  const segmentCountBuffer = convertIntToBuffer(segmentCount);
  headerBuffer.set(Buffer.concat([sequenceNumBuffer, segmentCountBuffer]));
  return Buffer.concat([headerBuffer, payload]);
}

function convertIntToBuffer(int: number): Buffer {
  const bytes = Buffer.alloc(4);
  bytes.writeIntLE(int, 0, 4);
  return bytes;
}

export default { createSegments };
