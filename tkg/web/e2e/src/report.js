function showhide(id) {
    var e = document.getElementById(id);
    e.style.display = (e.style.display === "block") ? "none" : "block";
}

function buildQuickLinks() {
    var failedSpecs = document.querySelectorAll("li.failed");
    var quickLinksContainer = document.getElementById("quickLinks");
    if (!quickLinksContainer) return;
    if (failedSpecs.length > 0) {
        document.getElementById("quickLinksHeader").textContent = "Quicklink of Failure"
    }
    for (var i = 0; i < failedSpecs.length; ++i) {
        var li = document.createElement("li");
        var a = document.createElement("a");
        a.href = "#" + failedSpecs[i].id;
        a.textContent = failedSpecs[i].dataset.name + "  (" + failedSpecs[i].dataset.browser + ")";
        li.appendChild(a);
        quickLinksContainer.appendChild(li);
    }
}

function updatePassCount() {
    var totalPassed = document.querySelectorAll("li.passed").length;
    var totalFailed = document.querySelectorAll("li.failed").length;
    var totalSpecs = totalFailed + totalPassed;
    console.log("passed: %s, failed: %s, total: %s", totalPassed, totalFailed, totalSpecs);
    document.getElementById("summaryTotalSpecs").textContent = document.getElementById("summaryTotalSpecs").textContent + totalSpecs;
    document.getElementById("summaryTotalFailed").textContent = document.getElementById("summaryTotalFailed").textContent + totalFailed;
    if (totalFailed) {
        document.getElementById("summary").className = "failed";
    }
}

function start() {
    updatePassCount();
    buildQuickLinks();
}
window.onload = start;