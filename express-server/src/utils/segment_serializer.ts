import { Channel, ConsumeMessage } from "amqplib";
import { SEARCH_QUEUE_CB } from "../rabbitmq/routing_keys";
import { rejects } from "assert";
import { resolve } from "path";

type SegmentHeader = {
  TotalSegments: number;
  SequenceNumber: number;
};

type WebpageBuffer = Buffer;

async function listenIncomingSegments(
  channel: Channel,
  generator: () => AsyncGenerator<any>,
): Promise<Buffer> {
  let expectedSequenceNum = 0;
  let segmentCount = 0;
  let webpageBuffer: Uint8Array[] = [];

  // Retrieves new incoming segments
  for await (const msg of generator()) {
    const m = msg as unknown as {
      data: ConsumeMessage;
      err: Error | null;
    };

    if (m.err != null) {
      throw new Error("Error while receiving segments");
    }
    const segment = decodeSegments(m.data.content);

    if (segment.header.SequenceNumber !== expectedSequenceNum) {
      channel!.nackAll(false);
      throw new Error("Unexpected sequence number");
    }

    webpageBuffer.push(segment.payload);
    channel!.ack(m.data, false);
    expectedSequenceNum++;
    segmentCount++;

    if (segmentCount == segment.header.TotalSegments) {
      console.log("Receieved all segments from search engine");
      break;
    }
  }

  console.log("Done");

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

function parseWebpages(webpageBuffer: Buffer): Array<{
  Url: string;
  Contents: string;
  Title: string;
}> {
  try {
    const jsonParse = JSON.parse(webpageBuffer.toString());
    return jsonParse;
  } catch (err) {
    const error = err as Error;
    console.log("Error: Unable to parse webpage buffer to json");
    console.error(error.message);
    return [];
  }
}

export default {
  listenIncomingSegments,
  decodeSegments,
  parseWebpages,
  getSegmentHeader,
};
