import zlib from "zlib";
export async function CompressData(
  data: string | Buffer | ArrayBufferLike,
  level: number,
) {
  const compressed = zlib.deflateSync(data, { level });
  return compressed;
}

export async function DecompressData(data: Buffer) {
  const decompressed = zlib.inflateSync(data);
  return decompressed;
}
