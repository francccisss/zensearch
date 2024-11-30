import zlib from "zlib";
const compressionOptions = { level: zlib.constants.Z_BEST_COMPRESSION };
export async function CompressData(data: any) {
  const compressed = zlib.deflateSync(data, compressionOptions);
  return compressed;
}

export async function DecompressData(data: any) {
  const decompressed = zlib.inflateSync(data);
  return decompressed;
}
