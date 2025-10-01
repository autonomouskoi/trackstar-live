import { Controller, setCB } from "./controller.js";
import { Current } from "./current.js";
import { SetsList } from "./sets.js";
import { TrackList } from "./tracklist.js";

function start() {
    let match = window.location.pathname.match(new RegExp("^/u/(?<userID>[A-Za-z0-9-]+)(?:/(?<setID>[0-9]+))?"));
    if (!match || !match.groups['userID']) {
        document.body.innerHTML = `WAH?`;
        return;
    }
    let userID = match.groups['userID'];
    document.querySelector('section.header h1').innerHTML = userID;

    let setID = 0;
    try {
        setID = Number(match.groups['setID']);
        if (!setID) {
            setID = 0;
        }
    } finally { }

    let goodSetCBs: setCB[] = new Array();
    let badSetCBs: setCB[] = new Array();

    // current
    let current = new Current();
    document.querySelector('section.header').appendChild(current);

    let tl = new TrackList();
    document.body.appendChild(tl);

    let h2 = document.querySelector('section.header h2');
    let ctrl = new Controller({
        userID,
        goodSet: (setID) => goodSetCBs.forEach((cb) => cb(setID)),
        badSet: (setID) => badSetCBs.forEach((cb) => cb(setID)),
        tracksLoaded: (updates) => tl.tracksLoaded(updates),
        newTrack: (update) => {
            tl.newTrack(update);
            current.newTrack(update);
        },
    });

    goodSetCBs.push((setID: number) => {
        h2.innerHTML = `
${new Date(setID).toLocaleString()}
&nbsp; <a href="/_trackUpdate/${userID}/${setID}?download=csv" class="button-link">CSV⇩</a>
&nbsp; <a href="/_trackUpdate/${userID}/${setID}" class="button-link" target="_main">JSON⇩</a>
`;
    });
    badSetCBs.push((setID: number) => { h2.innerHTML = `Set not found: ${setID}` });

    // sets
    let getSets = ctrl.getSets();
    let setsList = new SetsList(getSets, (setID) => ctrl.selectSet(setID));
    document.querySelector('nav#sets-list').appendChild(setsList);
    goodSetCBs.push((setID) => setsList.selectSet(setID));
    badSetCBs.push((setID) => setsList.selectSet(setID));

    ctrl.selectSet(setID);
}

start();