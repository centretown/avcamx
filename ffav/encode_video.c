#include <stdio.h>
#include <stdlib.h>

#include "encode_video.h"

static int internal_encode(AVCodecContext *enc_ctx, AVFrame *frame,
                           AVPacket *pkt, FILE *outfile) {
  int ret = 0;

  /* send the frame to the encoder */
  // if (frame)
  //   printf("Send frame %3" PRId64 "\n", frame->pts);

  ret = avcodec_send_frame(enc_ctx, frame);
  if (ret < 0) {
    fprintf(stderr, "Error sending a frame for encoding\n");
    return SEND_FRAME_FAILED;
  }

  while (ret >= 0) {
    ret = avcodec_receive_packet(enc_ctx, pkt);
    if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF)
      return SUCCESS;
    else if (ret < 0) {
      fprintf(stderr, "Error during encoding\n");
      return RECEIVE_PACKET_FAILED;
    }

    // printf("Write packet %3" PRId64 " (size=%5d)\n", pkt->pts, pkt->size);
    fwrite(pkt->data, 1, pkt->size, outfile);
    av_packet_unref(pkt);
  }
  return SUCCESS;
}

int encode_file(char *filename, char *codec_name) {
  // const char *filename, *codec_name;
  const AVCodec *codec;
  AVCodecContext *context = NULL;
  int i, ret, x, y;
  FILE *fileout;
  AVFrame *frame;
  AVPacket *packet;
  uint8_t endcode[] = {0, 0, 1, 0xb7};

  // if (argc <= 2) {
  //   fprintf(stderr, "Usage: %s <output file> <codec name>\n", argv[0]);
  //   return(0);
  // }
  // filename = argv[1];
  // codec_name = argv[2];

  /* find the mpeg1video encoder */
  codec = avcodec_find_encoder_by_name(codec_name);
  if (!codec) {
    fprintf(stderr, "Codec '%s' not found\n", codec_name);
    return FIND_CODEC_FAILED;
  }

  context = avcodec_alloc_context3(codec);
  if (!context) {
    fprintf(stderr, "Could not allocate video codec context\n");
    return ALLOCATE_CONTEXT_FAILED;
  }

  packet = av_packet_alloc();
  if (!packet) {
    fprintf(stderr, "Could not allocate av packet context\n");
    return ALLOCATE_PACKET_FAILED;
  }

  /* put sample parameters */
  context->bit_rate = 400000;
  /* resolution must be a multiple of two */
  context->width = 352;
  context->height = 288;
  /* frames per second */
  context->time_base = (AVRational){1, 25};
  context->framerate = (AVRational){25, 1};

  /* emit one intra frame every ten frames
   * check frame pict_type before passing frame
   * to encoder, if frame->pict_type is AV_PICTURE_TYPE_I
   * then gop_size is ignored and the output of encoder
   * will always be I frame irrespective to gop_size
   */
  context->gop_size = 10;
  context->max_b_frames = 1;
  context->pix_fmt = AV_PIX_FMT_YUV420P;

  if (codec->id == AV_CODEC_ID_H264)
    av_opt_set(context->priv_data, "preset", "slow", 0);

  /* open it */
  ret = avcodec_open2(context, codec, NULL);
  if (ret < 0) {
    fprintf(stderr, "Could not open codec: %s\n", av_err2str(ret));
    return CODEC_OPEN_FAILED;
  }

  fileout = fopen(filename, "wb");
  if (!fileout) {
    fprintf(stderr, "Could not open %s\n", filename);
    return FILE_OPEN_FAILED;
  }

  frame = av_frame_alloc();
  if (!frame) {
    fprintf(stderr, "Could not allocate video frame\n");
    return ALLOCATE_FRAME_FAILED;
  }
  frame->format = context->pix_fmt;
  frame->width = context->width;
  frame->height = context->height;

  ret = av_frame_get_buffer(frame, 0);
  if (ret < 0) {
    fprintf(stderr, "Could not allocate the video frame data\n");
    return ALLOCATE_FRAME_DATA_FAILED;
  }

  /* encode 1 second of video */
  for (i = 0; i < 200; i++) {
    fflush(stdout);

    /* Make sure the frame data is writable.
       On the first round, the frame is fresh from av_frame_get_buffer()
       and therefore we know it is writable.
       But on the next rounds, encode() will have called
       avcodec_send_frame(), and the codec may have kept a reference to
       the frame in its internal structures, that makes the frame
       unwritable.
       av_frame_make_writable() checks that and allocates a new buffer
       for the frame only if necessary.
     */
    ret = av_frame_make_writable(frame);
    if (ret < 0)
      return WRITE_FRAME_FAILED;

    /* Prepare a dummy image.
       In real code, this is where you would have your own logic for
       filling the frame. FFmpeg does not care what you put in the
       frame.
     */
    /* Y */
    for (y = 0; y < context->height; y++) {
      for (x = 0; x < context->width; x++) {
        frame->data[0][y * frame->linesize[0] + x] = x + y + i * 3;
      }
    }

    /* Cb and Cr */
    for (y = 0; y < context->height / 2; y++) {
      for (x = 0; x < context->width / 2; x++) {
        frame->data[1][y * frame->linesize[1] + x] = 128 + y + i * 2;
        frame->data[2][y * frame->linesize[2] + x] = 64 + x + i * 5;
      }
    }

    frame->pts = i;

    /* encode the image */
    internal_encode(context, frame, packet, fileout);
  }

  /* flush the encoder */
  internal_encode(context, NULL, packet, fileout);

  /* Add sequence end code to have a real MPEG file.
     It makes only sense because this tiny examples writes packets
     directly. This is called "elementary stream" and only works for some
     codecs. To create a valid file, you usually need to write packets
     into a proper file format or protocol; see mux.c.
   */
  if (codec->id == AV_CODEC_ID_MPEG1VIDEO ||
      codec->id == AV_CODEC_ID_MPEG2VIDEO)
    fwrite(endcode, 1, sizeof(endcode), fileout);
  fclose(fileout);

  avcodec_free_context(&context);
  av_frame_free(&frame);
  av_packet_free(&packet);

  return 0;
}

void initContext(AVCodecContext *context) {
  /* put sample parameters */
  context->bit_rate = 400000;
  /* resolution must be a multiple of two */
  context->width = 352;
  context->height = 288;
  /* frames per second */
  context->time_base = (AVRational){1, 25};
  context->framerate = (AVRational){25, 1};

  /* emit one intra frame every ten frames
   * check frame pict_type before passing frame
   * to encoder, if frame->pict_type is AV_PICTURE_TYPE_I
   * then gop_size is ignored and the output of encoder
   * will always be I frame irrespective to gop_size
   */
  context->gop_size = 10;
  context->max_b_frames = 1;
  context->pix_fmt = AV_PIX_FMT_YUV420P;
}

EncodeData *encode_init(char *codec_name, char *filename) {
  EncodeData *data = calloc(1, sizeof(EncodeData));
  data->codec = avcodec_find_encoder_by_name(codec_name);
  if (!data->codec) {
    fprintf(stderr, "Codec '%s' not found\n", codec_name);
    data->ret = FIND_CODEC_FAILED;
    return data;
  }

  data->context = avcodec_alloc_context3(data->codec);
  if (!data->context) {
    fprintf(stderr, "Could not allocate video codec context\n");
    data->ret = ALLOCATE_CONTEXT_FAILED;
    return data;
  }

  data->packet = av_packet_alloc();
  if (!data->packet) {
    fprintf(stderr, "Could not allocate av packet context\n");
    data->ret = ALLOCATE_PACKET_FAILED;
    return data;
  }

  initContext(data->context);

  if (data->codec->id == AV_CODEC_ID_H264) {
    av_opt_set(data->context->priv_data, "preset", "slow", 0);
  }

  int ret = avcodec_open2(data->context, data->codec, NULL);
  if (ret < 0) {
    fprintf(stderr, "Could not open codec: %s\n", av_err2str(ret));
    data->ret = CODEC_OPEN_FAILED;
    return data;
  }

  data->fileout = fopen(filename, "wb");
  if (!data->fileout) {
    fprintf(stderr, "Could not open %s\n", filename);
    data->ret = FILE_OPEN_FAILED;
    return data;
  }

  data->frame = av_frame_alloc();
  if (!data->frame) {
    fprintf(stderr, "Could not allocate video frame\n");
    data->ret = ALLOCATE_FRAME_FAILED;
    return data;
  }

  data->frame->format = data->context->pix_fmt;
  data->frame->width = data->context->width;
  data->frame->height = data->context->height;

  ret = av_frame_get_buffer(data->frame, 0);
  if (ret < 0) {
    fprintf(stderr, "Could not allocate the video frame data\n");
    data->ret = ALLOCATE_FRAME_DATA_FAILED;
    return data;
  }

  return data;
}

void encode_free(EncodeData *data) {

  if (data->context != NULL && data->packet != NULL && data->fileout != NULL) {
    fprintf(stderr, "encode final\n\n");
    internal_encode(data->context, NULL, data->packet, data->fileout);
  }

  if (data->context != NULL) {
    avcodec_free_context(&data->context);
    data->context = NULL;
  }

  if (data->frame != NULL) {
    av_frame_free(&data->frame);
    data->frame = NULL;
  }

  if (data->packet != NULL) {
    av_packet_free(&data->packet);
    data->packet = NULL;
  }

  if (data->fileout != NULL) {
    if (data->codec->id == AV_CODEC_ID_MPEG1VIDEO ||
        data->codec->id == AV_CODEC_ID_MPEG2VIDEO) {
      uint8_t endcode[] = {0, 0, 1, 0xb7};
      fwrite(endcode, 1, sizeof(endcode), data->fileout);
    }
    fclose(data->fileout);
    data->fileout = NULL;
  }

  free(data);
}

void encode_write_data(EncodeData *data, int index) {
  fflush(stdout);

  /* Make sure the frame data is writable.
     On the first round, the frame is fresh from av_frame_get_buffer()
     and therefore we know it is writable.
     But on the next rounds, encode() will have called
     avcodec_send_frame(), and the codec may have kept a reference to
     the frame in its internal structures, that makes the frame
     unwritable.
     av_frame_make_writable() checks that and allocates a new buffer
     for the frame only if necessary.
   */
  int ret = av_frame_make_writable(data->frame);
  if (ret < 0) {
    data->ret = WRITE_FRAME_FAILED;
    return;
  }

  /* Prepare a dummy image.
     In real code, this is where you would have your own logic for
     filling the frame. FFmpeg does not care what you put in the
     frame.
   */
  /* Y */
  AVFrame *frame = data->frame;
  for (int y = 0; y < data->context->height; y++) {
    for (int x = 0; x < data->context->width; x++) {
      frame->data[0][y * frame->linesize[0] + x] = x + y + index * 3;
    }
  }

  /* Cb and Cr */
  for (int y = 0; y < data->context->height / 2; y++) {
    for (int x = 0; x < data->context->width / 2; x++) {
      frame->data[1][y * frame->linesize[1] + x] = 128 + y + index * 2;
      frame->data[2][y * frame->linesize[2] + x] = 64 + x + index * 5;
    }
  }

  frame->pts = index;

  /* encode the image */
  data->ret =
      internal_encode(data->context, frame, data->packet, data->fileout);
}

void encode_bytes(EncodeData *data, int x, int y, char cY, char cCb, char cCy) {
  AVFrame *frame = data->frame;
  frame->data[0][y * frame->linesize[0] + x] = cY;
  frame->data[1][y * frame->linesize[1] + x] = cCb;
  frame->data[2][y * frame->linesize[2] + x] = cCy;
}

void encode_buffer(EncodeData *data, int index, unsigned char *bufY,
                   unsigned char *bufCb, unsigned char *bufCr) {
  fflush(stdout);

  int ret = av_frame_make_writable(data->frame);
  if (ret < 0) {
    data->ret = WRITE_FRAME_FAILED;
    return;
  }

  /* Prepare a dummy image.
     In real code, this is where you would have your own logic for
     filling the frame. FFmpeg does not care what you put in the
     frame.
   */
  /* Y */
  AVFrame *frame = data->frame;
  int offs = 0;
  for (int y = 0; y < data->context->height; y++) {
    for (int x = 0; x < data->context->width; x++) {
      offs = y * frame->linesize[0] + x;
      frame->data[0][offs] = bufY[y * data->context->width + x];
    }
  }

  /* Cb and Cr */
  for (int y = 0; y < data->context->height / 2; y++) {
    for (int x = 0; x < data->context->width / 2; x++) {
      offs = y * frame->linesize[1] + x;
      frame->data[1][offs] = bufCb[y * data->context->width + x];
      offs = y * frame->linesize[2] + x;
      frame->data[2][offs] = bufCr[y * data->context->width + x];
      // frame->data[2][y * frame->linesize[2] + x] = 64 + x + index * 5;
    }
  }

  frame->pts = index;

  /* encode the image */
  data->ret =
      internal_encode(data->context, frame, data->packet, data->fileout);
}

void encode(EncodeData *data, int index) {
  data->frame->pts = index;
  data->ret =
      internal_encode(data->context, data->frame, data->packet, data->fileout);
}

void encode_write_prepare(EncodeData *data) {
  fflush(stdout);
  int ret = av_frame_make_writable(data->frame);
  if (ret < 0) {
    data->ret = WRITE_FRAME_FAILED;
    return;
  }
}
