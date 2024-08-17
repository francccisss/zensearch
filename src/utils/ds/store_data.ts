import { data_t } from "../../types/data_t";
import BTree from "./btree"
import fs from "fs"

(async function(path:string){
    const btree = new BTree(30);
    const file = fs.readFileSync(path,"utf-8");
    const to_json :data_t= JSON.parse(file);
    to_json.webpage_contents.forEach((content,i)=>{
        btree.insert_and_split(btree.root,{content,key:i})
    })
    fs.writeFileSync("./bstree.json",JSON.stringify(btree.root))

})(process.argv[2])