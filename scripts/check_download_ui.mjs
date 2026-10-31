import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import vm from "node:vm";
const source=readFileSync(new URL("../internal/echo/web/app.js",import.meta.url),"utf8");
const start=source.indexOf("function projects(){"),end=source.indexOf("function overviewView(){",start);
assert.ok(start>=0&&end>start);
const context={
 state:{data:{report:{report:{sources:[{work_id:1,last_comparable_changes:{from:"2026-10-30",to:"2026-10-31",known_downloads:{total:12,packages:10},unknown_assets:2},latest:{downloads:{total:49,packages:40,source:2,auxiliary:7,other:0},assets:[{name:"app.exe",category:"package",download_count:40}]}}]}}}},
 pref:()=>"all",settingsLabel:pair=>pair[1],T:value=>value,E:value=>String(value??""),fmt:value=>value==null?"—":String(value),dt:value=>String(value)
};
vm.createContext(context);vm.runInContext(source.slice(start,end),context);
const project=context.projects()[0],html=context.downloadDetails(project);
assert.equal(project.downloads,49);
assert.equal(project.changes.unknown_assets,2);
assert.ok(html.includes("+12")&&html.includes("+10"),"UI must use report.last_comparable_changes, not a missing changes field");
assert.ok(html.includes("2026-10-30")&&html.includes("2026-10-31"));
assert.ok(!html.includes("Two complete release observations"));
context.state.data.report.report.sources[0].latest.downloads=null;
assert.equal(context.projects()[0].downloads,null,"Missing downloads must remain unknown");
let handler;let renders=0;
const hashContext={window:{addEventListener:(name,fn)=>{assert.equal(name,"hashchange");handler=fn;}},location:{hash:"#works"},nav:[["overview"],["works"]],state:{view:"overview"},render:()=>renders++};
vm.runInNewContext(source.slice(source.lastIndexOf('window.addEventListener("hashchange"')),hashContext);
handler();assert.equal(hashContext.state.view,"works");assert.equal(renders,1);
handler();assert.equal(renders,1,"Unchanged hash does not render twice");
hashContext.location.hash="#unknown";handler();assert.equal(hashContext.state.view,"works");
console.log("Download interval UI contract and history navigation passed.");
