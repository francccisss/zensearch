import test, { describe, mock } from "node:test";
import segmentSerializer from "../serializer/segment_serializer.ts";
import { mockWebpages } from "./test_objects.ts";
import { fail } from "node:assert";
import { CompressData, DecompressData } from "../compression.ts";

describe("Creating segments", () => {
  let dataL = 10000;
  let generatedWebpages = [];
  for (let i = 0; i < dataL; i++) {
    generatedWebpages.push(mockWebpages);
  }
  test("Data segmentation with compression", async () => {
    try {
      const compressed = await CompressData(JSON.stringify(generatedWebpages));

      const webpageBuf = segmentSerializer.createSegments(
        Buffer.from(compressed),
        100000,
      );

      if (webpageBuf.length == 0) {
        throw new Error("Length of 0");
      }
      console.log("Length while compressed: %d bytes", compressed.byteLength);
      console.log("Total Segments created %d", webpageBuf.length);
    } catch (err) {
      console.error((err as Error).message);
      fail("Failed to create segments");
    }
  });

  test("Data segmentation without compression", async () => {
    try {
      const bufInput = Buffer.from(JSON.stringify(generatedWebpages));
      const webpageBuf = segmentSerializer.createSegments(bufInput, 100000);
      if (webpageBuf.length == 0) {
        throw new Error("Length of 0");
      }
      console.log("Length while not compressed: %d bytes", bufInput.byteLength);
      console.log("Total Segments created %d", webpageBuf.length);
    } catch (err) {
      console.error((err as Error).message);
      fail("Failed to create segments");
    }
  });
});
