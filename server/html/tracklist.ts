import type { TrackUpdate } from "/tspb/trackstar_pb.js";

class TrackList extends HTMLTableElement {
    private _lastIdx = 0;

    constructor() {
        super();

        this._reset();
    }

    private _reset() {
        this.innerHTML = `
<thead>
    <th>#</th>
    <th>Artist</th>
    <th>Title</th>
    <th>When</th>
    <th>Deck ID</th>
</thead>
`;
    }

    tracksLoaded(updates: TrackUpdate[]) {
        this._reset();
        updates.forEach((tu) => this.newTrack(tu));
    }

    newTrack(tu: TrackUpdate) {
        if (tu.index == this._lastIdx) {
            return;
        }
        let tr = document.createElement('tr');
        addTD(tr, tu.index.toString());
        addTD(tr, tu.track.artist);
        addTD(tr, tu.track.title);
        addTD(tr, new Date(Number(tu.when) * 1000).toLocaleTimeString());
        addTD(tr, tu.deckId ? tu.deckId : '');
        this.appendChild(tr);
        this._lastIdx = tu.index;
    }
}
customElements.define('tslive-tracklist', TrackList, { extends: 'table' });

function addTD(parent: HTMLElement, value: string) {
    let d = document.createElement('td');
    d.innerText = value;
    parent.appendChild(d);
}

export { TrackList };