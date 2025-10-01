import type { TrackUpdate } from "/tspb/trackstar_pb.js"

class Current extends HTMLDivElement {
    private _idx: HTMLDivElement;
    private _artist: HTMLDivElement;
    private _title: HTMLDivElement;
    private _deckID: HTMLDivElement;

    constructor() {
        super();
        this.classList.add('current');
    }

    newTrack(tu: TrackUpdate) {
        if (!this._idx) {
            this._idx = this._addField('#');
            this._artist = this._addField('Artist');
            this._title = this._addField('Title');
            this._deckID = this._addField('Deck ID');
        }
        this.classList.remove('fadeIn');
        this.classList.add('fadeOut');
        this.addEventListener('animationend', () => {
            this._idx.innerText = tu.index.toString();
            this._artist.innerText = tu.track.artist;
            this._title.innerText = tu.track.title;
            this._deckID.innerText = tu.deckId ? tu.deckId : '';
            this.classList.remove('fadeOut');
            this.classList.add('fadeIn');
        }, { once: true });
    }

    private _addField(labelText: string): HTMLDivElement {
        let label = document.createElement('div');
        label.classList.add('column-header');
        label.innerText = labelText;
        this.appendChild(label);

        let v = document.createElement('div');
        this.appendChild(v);
        return v;
    }
}
customElements.define('tslive-current', Current, { extends: 'div' });

export { Current };