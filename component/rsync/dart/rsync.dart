import 'package:typed_data/typed_data.dart' as typed;
import 'dart:typed_data';
import 'dart:convert';
import 'dart:convert' show utf8;

import 'crclib.dart';

const int DeltaMagic = 0x7273;
const int RS_OP_END  =0x00;	//结束
const int RS_OP_BLOCK_N1 	= 0x01;	//匹配块index 1字节 0-256 个BLOCK内的数据有效
const int RS_OP_BLOCK_N2 	= 0x02;
const int RS_OP_BLOCK_N4 	= 0x03;
const int RS_OP_DATA_N1 	= 0x04;	//未匹配块  数据长度 0-256
const int RS_OP_DATA_N2 	= 0x05;
const int RS_OP_DATA_N4 	= 0x06;

void _copy( ByteData source,int srcPos, ByteData dest, int desPos,int length) {
  //逻辑代码；
  for(var i =0; i <length;i++)
  {
    dest.setUint8(desPos+i, source.getUint8(srcPos+i));
    //dest[desPos+i] = source[srcPos+i];
  }
}

ByteData Utf8ToByteData(String data){

  var outputAsUint8List = new Uint8List.fromList(utf8.encode(data));
  return ByteData.view(outputAsUint8List.buffer);
}

String ByteDataToUtf8(ByteData data){
  return utf8.decode(data.buffer.asUint8List());
}

ByteData RsyncPatch(ByteData patchdata, ByteData content) {
  var offset=0,retoffset=0;
  var deltamagic = patchdata.getInt16(offset);
  if(DeltaMagic!=deltamagic){
    throw   "不是rsync patch消息头";
  }
  offset+=2;
  var pcrc32 = patchdata.getUint32(offset);
  offset+=4;
  var BlockSize = patchdata.getUint32(offset);
  offset+=4;
  var modifiedSize = patchdata.getUint32(offset);
  offset+=4;
  var result= new ByteData(modifiedSize);

  do
  {
    int cmd= patchdata.getUint8(offset);
    offset+=1;
    if (cmd == RS_OP_END) { // delta的结束命令
      break;
    }
    if (patchdata.lengthInBytes <= offset) { // 超出了
      throw   "数据解析异常";
    }
    switch (cmd) {
      case RS_OP_BLOCK_N1:
        var start_index=patchdata.getUint8(offset);
        offset+=1;
        var end_index=patchdata.getUint8(offset);
        offset+=1;
        for (var i=start_index;i<=end_index;i++){
          var desPos=i*BlockSize;
          _copy(content,desPos,result,retoffset, BlockSize);
          retoffset += BlockSize;
        }
        break;
      case RS_OP_BLOCK_N2:
        var start_index=patchdata.getUint16(offset);
        offset+=2;
        var end_index=patchdata.getUint16(offset);
        offset+=2;
        for (var i=start_index;i<=end_index;i++){
          var desPos=i*BlockSize;
          _copy(content,desPos,result,retoffset, BlockSize);
          retoffset += BlockSize;
        }
        break;
      case RS_OP_BLOCK_N4:
        var start_index=patchdata.getUint32(offset);
        offset+=4;
        var end_index=patchdata.getUint32(offset);
        offset+=4;
        for (var i=start_index;i<=end_index;i++){
          var desPos=i*BlockSize;
          _copy(content,desPos,result,retoffset, BlockSize);
          retoffset += BlockSize;
        }
        break;
      case RS_OP_DATA_N1:
        var lenght=patchdata.getUint8(offset);
        offset+=1;
        //读数据
        _copy(patchdata,offset,result,retoffset, lenght);
        retoffset += lenght;
        offset+=lenght;
        break;
      case RS_OP_DATA_N2:
        var lenght=patchdata.getUint16(offset);
        offset+=2;
        //读数据
        _copy(patchdata,offset,result,retoffset, lenght);
        retoffset += lenght;
        offset+=lenght;
        break;
      case RS_OP_DATA_N4:
        var lenght=patchdata.getUint32(offset);
        offset+=4;
        //读数据
        _copy(patchdata,offset,result,retoffset, lenght);
        retoffset += lenght;
        offset+=lenght;
        break;
      default:
        throw   "解析到未知指令 ${cmd}";
    }
  }
  while (true);
  var rcrc32=new Crc32Zlib().convert(result.buffer.asUint8List());
  //console.log(patchDocument.length,offset,rcrc32);
  if (rcrc32!=pcrc32){
    throw   "数据结果校验失败";
  }
  return result;
}