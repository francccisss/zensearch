type SegmentHeader = {
  TotalSegments: number;
  SequenceNumber: number;
};

//
// MSS is number in bytes
function createSegments(
  webpages: Buffer, // webpages queried from database
  MSS: number,
  handler?: (newSegment: Buffer) => Promise<void>,
): Array<Buffer> {
  const dataLength = webpages.byteLength;
  let currentIndex = 0;
  let segmentCount = Math.trunc(dataLength / MSS) + 1; // + 1 to store the remainder
  let segments: Array<Buffer> = [];
  let pointerPosition = MSS;

  for (let i = 0; i < segmentCount; i++) {
    let slicedArray = webpages.slice(currentIndex, pointerPosition);

    currentIndex += slicedArray.byteLength;
    // Add to offset MSS to point to the next segment in the array
    // manipulate pointerPosition to adjust to lower values using Math.min()

    // Is current data length enough to fit MSS?
    // if so add from current position + MSS
    // else get remaining of the currentDataLength
    pointerPosition += Math.min(MSS, Math.abs(currentIndex - dataLength));
    const payload = Buffer.alloc(slicedArray.length);
    payload.set(slicedArray);
    const segment = newSegment(i, segmentCount, Buffer.from(payload));

    if (handler !== undefined) {
      handler(segment);
    }
    segments.push(segment);
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

// FOR TESTING
function listenIncomingSegments(segments: Array<Buffer>): Buffer {
  let expectedSequenceNum = 0;
  let segmentCount = 0;
  let webpageBuffer: Uint8Array[] = [];

  for (let i = 0; i < segments.length; i++) {
    const segment = decodeSegments(segments[i]);
    if (segment.header.SequenceNumber !== expectedSequenceNum) {
      throw new Error("Unexpected sequence number");
    }

    webpageBuffer.push(segment.payload);
    expectedSequenceNum++;
    segmentCount++;

    if (segmentCount == segment.header.TotalSegments) {
      console.log("Receieved all segments from search engine");
      console.log("Total Segments Decoded: %d", segmentCount);
      expectedSequenceNum = 0;
      segmentCount = 0;
      break;
    }
  }

  return Buffer.concat(webpageBuffer);
}

function decodeSegments(segment: Buffer): {
  header: SegmentHeader;
  payload: Uint8Array;
} {
  return {
    header: getSegmentHeader(segment.slice(0, 8)),
    payload: getSegmentPayload(segment),
  };
}

function getSegmentHeader(bytes: Buffer): SegmentHeader {
  const seqNumBuff = bytes.slice(0, 4);
  const totalSegmentsBuff = bytes.slice(4, 8);
  return {
    SequenceNumber: seqNumBuff.readUint32LE(),
    TotalSegments: totalSegmentsBuff.readUint32LE(),
  };
}

function getSegmentPayload(bytes: Buffer): Buffer {
  return bytes.slice(8);
}
export default { createSegments, listenIncomingSegments };
