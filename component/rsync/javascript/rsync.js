var DeltaMagic = 0x7273;
var RS_OP_END  =0x00;	//结束
var RS_OP_BLOCK_N1 	= 0x01;	//匹配块index 1字节 0-256 个BLOCK内的数据有效
var RS_OP_BLOCK_N2 	= 0x02;
var RS_OP_BLOCK_N4 	= 0x03;
var RS_OP_DATA_N1 	= 0x04;	//未匹配块  数据长度 0-256
var RS_OP_DATA_N2 	= 0x05;
var RS_OP_DATA_N4 	= 0x06;
/**
 * This is a function born of annoyance. You can't create a Uint32Array at a non 4 byte boundary. So this is necessary to read
 * a 32bit int from an arbitrary location. Lame.
 *
 * BTW: This assumes everything to be big endian.
 */
function readUint32(uint8View, offset)
{
    return (uint8View[offset]<<24 | uint8View[++offset] << 16 | uint8View[++offset] << 8 | uint8View[++offset]) >>> 0;
}
function readUint16(uint8View, offset)
{
    return (uint8View[offset] << 8 | uint8View[++offset]) >>> 0;
}
function readUint8(uint8View, offset)
{
    //console.log("readUint8",uint8View[offset]);
    return (uint8View[offset]) >>> 0;
}

function copy( source,srcPos, dest, desPos,length) {
    //逻辑代码；
    for(var i =0; i <length;i++)
    {
        dest[desPos+i] = source[srcPos+i];
    }
}

/**
 * Convert an Uint8Array into a string.
 *
 * @returns {String}
 */
function Decodeuint8arr(uint8array){
    return new TextDecoder("utf-8").decode(uint8array);
}

/**
 * Convert a string into a Uint8Array.
 *
 * @returns {Uint8Array}
 */
function Encodeuint8arr(myString){
    return new TextEncoder("utf-8").encode(myString);
}

function RsyncPatch(patchDocument, content) {
    var offset=0,retoffset=0;
    var deltamagic = readUint16(patchDocument,offset);
    if(DeltaMagic!=deltamagic){
        throw   "不是rsync patch消息头";
    }
    offset+=2;
    var pcrc32 = readUint32(patchDocument,offset);
    offset+=4;
    var BlockSize = readUint32(patchDocument,offset);
    offset+=4;
    var modifiedSize = readUint32(patchDocument,offset);
    offset+=4;
    var result= new Uint8Array(modifiedSize);
    do
    {
        var cmd=readUint8(patchDocument,offset);
        offset+=1;
        if (cmd === RS_OP_END) { // delta的结束命令
            break
        }
        if (patchDocument.length <= offset) { // 超出了
            throw   "数据解析异常";
        }
        switch (cmd) {
            case RS_OP_BLOCK_N1:
                var start_index=readUint8(patchDocument,offset);
                offset+=1;
                var end_index=readUint8(patchDocument,offset);
                offset+=1;
                for (var i=start_index;i<=end_index;i++){
                    var desPos=i*BlockSize;
                    copy(content,desPos,result,retoffset, BlockSize);
                    retoffset += BlockSize;
                }
                break
            case RS_OP_BLOCK_N2:
                var start_index=readUint16(patchDocument,offset);
                offset+=2;
                var end_index=readUint16(patchDocument,offset);
                offset+=2;
                for (var i=start_index;i<=end_index;i++){
                    var desPos=i*BlockSize;
                    copy(content,desPos,result,retoffset, BlockSize);
                    retoffset += BlockSize;
                }
                break
            case RS_OP_BLOCK_N4:
                var start_index=readUint32(patchDocument,offset);
                offset+=4;
                var end_index=readUint32(patchDocument,offset);
                offset+=4;
                for (var i=start_index;i<=end_index;i++){
                    var desPos=i*BlockSize;
                    copy(content,desPos,result,retoffset, BlockSize);
                    retoffset += BlockSize;
                }
                break
            case RS_OP_DATA_N1:
                var lenght=readUint8(patchDocument,offset);
                offset+=1;
                //读数据
                copy(patchDocument,offset,result,retoffset, lenght);
                retoffset += lenght;
                offset+=lenght;
                break
            case RS_OP_DATA_N2:
                var lenght=readUint16(patchDocument,offset);
                offset+=2;
                //读数据
                copy(patchDocument,offset,result,retoffset, lenght);
                retoffset += lenght;
                offset+=lenght;
                break
            case RS_OP_DATA_N4:
                var lenght=readUint16(patchDocument,offset);
                offset+=4;
                //读数据
                copy(patchDocument,offset,result,retoffset, lenght);
                retoffset += lenght;
                offset+=lenght;
                break
            default:
                throw   "解析到未知指令";
        }
    }
    while (true);
    var rcrc32=crc32(result);
    //console.log(patchDocument.length,offset,rcrc32);
    if (rcrc32!=pcrc32){
        throw   "数据结果校验失败";
    }
    return result;
}

(function () {
    'use strict';

    var root = typeof window === 'object' ? window : {};
    var NODE_JS = !root.JS_CRC_NO_NODE_JS && typeof process === 'object' && process.versions && process.versions.node;
    if (NODE_JS) {
        root = global;
    }
    var COMMON_JS = !root.JS_CRC_NO_COMMON_JS && typeof module === 'object' && module.exports;
    var AMD = typeof define === 'function' && define.amd;


    var Modules = [
        {
            name: 'RsyncPatch',
            method:RsyncPatch
        }
    ];


    var exports = {};
    for (var i = 0;i < Modules.length;++i) {
        var m = Modules[i];
        exports[m.name] = m.method;
    }
    if (COMMON_JS) {
        module.exports = exports;
    } else {
        for (i = 0;i < Modules.length;++i) {
            var m = Modules[i];
            root[m.name] = m.method;
        }
        if (AMD) {
            define(function() {
                return exports;
            });
        }
    }
})();
