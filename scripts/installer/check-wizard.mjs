// Verify success/error/headless behavior using the actual shipped wizard scripts.
import fs from "node:fs";
import vm from "node:vm";
import assert from "node:assert/strict";
const root = new URL(".", import.meta.url);
function run({headless=false, success=true, app=true, extension=false}={}) {
  const pages=[], calls=[], hidden=[], signals=[];
  const form={openAppCheckBox:{checked:app},openExtensionCheckBox:{checked:extension}};
  const installer={
    isInstaller:()=>true, isCommandLineInstance:()=>headless, status:success?1:0,
    setDefaultPageVisible:(page,visible)=>hidden.push([page,visible]),
    installationFinished:{connect:()=>signals.push("installed")},finishButtonClicked:{connect:()=>signals.push("finish")},
    addWizardPageItem:()=>pages.push("finish"),value:()=>"C:/Reverfolio",
    environmentVariable:()=>"C:/Windows",toNativeSeparators:s=>s,
    executeDetached:(...args)=>calls.push(args)
  };
  const c={installer,component:{userInterface:()=>form},QInstaller:{Success:1,InstallationFinished:9,ComponentSelection:3,StartMenuSelection:5},console};
  vm.createContext(c);
  vm.runInContext(fs.readFileSync(new URL("controller.qs",root),"utf8"),c);
  vm.runInContext("new Controller()",c);
  vm.runInContext("function Component() {}",c);
  vm.runInContext(fs.readFileSync(new URL("finish.qs",root),"utf8").replaceAll("__PLATFORM__","windows"),c);
  vm.runInContext("Component.prototype.showFirstRunOptions(); Component.prototype.finishFirstRun()",c);
  return {hidden,pages,calls};
}
const normal=run();
assert.equal(normal.hidden.length,2);
assert.ok(normal.hidden.every(([,visible])=>visible===false));
assert.equal(normal.pages.length,1);
assert.equal(normal.calls.length,1);
assert.equal(normal.calls[0][0], "C:/Reverfolio/bin/reverfolio-host.exe");
assert.deepEqual(Array.from(normal.calls[0][1]), ["--open-ui"]);
assert.equal(run({app:false}).calls.length,0);
const ext=run({app:false,extension:true});
assert.equal(ext.calls.length,1);
assert.equal(ext.calls[0][0], "C:/Reverfolio/bin/reverfolio-host.exe");
assert.deepEqual(Array.from(ext.calls[0][1]), ["--guide-extension"]);
for(const options of [{headless:true},{success:false}]) {
  assert.equal(run(options).calls.length,0);
  assert.equal(run(options).pages.length,0);
}
console.log("Installer wizard: hidden single-product pages, optional Finish actions, no launch on headless/failure.");
