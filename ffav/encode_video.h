#ifndef ENCODE_VIDEO_H
#define ENCODE_VIDEO_H

#include <libavcodec/avcodec.h>
#include <libavutil/imgutils.h>
#include <libavutil/opt.h>

enum ErrorCode {
  SUCCESS,
  FIND_CODEC_FAILED,
  ALLOCATE_CONTEXT_FAILED,
  ALLOCATE_PACKET_FAILED,
  CODEC_OPEN_FAILED,
  FILE_OPEN_FAILED,
  ALLOCATE_FRAME_FAILED,
  ALLOCATE_FRAME_DATA_FAILED,
  WRITE_FRAME_FAILED,
  SEND_FRAME_FAILED,
  RECEIVE_PACKET_FAILED,
};

typedef struct Encoder {
  const AVCodec *codec;
  AVCodecContext *context;
} Encoder;

typedef struct EncodeData {
  const AVCodec *codec;
  AVCodecContext *context;
  FILE *fileout;
  AVFrame *frame;
  AVPacket *packet;
  int ret;
} EncodeData;

int encode_file(char *filename, char *codec_name);

EncodeData *encode_init(char *codec_name, char *filename);
void encode_write_data(EncodeData *data, int index);
void encode_free(EncodeData *data);
void encode_bytes(EncodeData *data, int x, int y, char cY, char cCb, char cCy);
void encode_write_prepare(EncodeData *data);
void encode(EncodeData *data, int index);
void encode_buffer(EncodeData *data, int index, unsigned char *bufY,
                   unsigned char *bufCb, unsigned char *bufCr);
#endif
