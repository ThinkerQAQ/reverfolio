// Adapted from VidOtter's successful-installation hooks; see third_party/vidotter.
Component.prototype.showFirstRunOptions = function() {
    if (installer.isInstaller() && !installer.isCommandLineInstance() &&
        installer.status == QInstaller.Success) {
        installer.addWizardPageItem(component, "ReverfolioFirstRun", QInstaller.InstallationFinished);
    }
};
Component.prototype.finishFirstRun = function() {
    if (!installer.isInstaller() || installer.isCommandLineInstance() ||
        installer.status != QInstaller.Success) return;
    var form = component.userInterface("ReverfolioFirstRun");
    if (!form) return;
    var target = installer.value("TargetDir");
    console.log("REVERFOLIO_FINISH_OPTIONS");
    if ("__PLATFORM__" == "windows") {
        var launcher = installer.toNativeSeparators(target + "/bin/reverfolio-host.exe");
        if (form.openAppCheckBox.checked) installer.executeDetached(launcher, ["--open-ui"], target);
        if (form.openExtensionCheckBox.checked) installer.executeDetached(launcher, ["--guide-extension"], target);
    } else {
        if (form.openAppCheckBox.checked) {
            installer.executeDetached("/bin/bash", [target + "/Open-Reverfolio.sh"], target);
        }
        if (form.openExtensionCheckBox.checked) {
            QDesktopServices.openUrl("file://" + target + "/extension");
            QDesktopServices.openUrl("file://" + target + "/README.md");
        }
    }
};
