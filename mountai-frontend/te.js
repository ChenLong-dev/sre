let a = [
    "a3434-3434-343",
    "ba_a",
    "c-aas3-2",
    "s_a3_2_545",
    "dsfdfrt_2434_343_3454",
    "dsfdfrt_2434_343_3454",
    "dsfdfrt-3m-343a-chen1",
]
    


 


// let reg = new RegExp(/^[_0-9a-zA-Z\_]+$/)  数字 英文 下划线
// let reg = new RegExp(/^[A-Za-z0-9_\( \)\-]+$/)  支持括号等
let reg = new RegExp(/^[A-Za-z0-9\-]+$/);  //支持  数字英文 中划线
// var reg = /^((\w*\d\w*[a-z]\w*)|(\w*[a-z]\w*\d\w*))$/i;
for(value of a){
    if(reg.test(value)) {
        console.log('value',value);
    }
}