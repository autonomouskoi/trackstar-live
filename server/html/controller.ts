import type { TrackUpdate } from "/tspb/trackstar_pb.js";

type setCB = (setID: number) => void;

interface ControllerArgs {
    userID: string,
    goodSet: setCB,
    badSet: setCB,
    tracksLoaded: (updates: TrackUpdate[]) => void,
    newTrack: (update: TrackUpdate) => void,
}

class Controller {
    private _user: string;
    private _sets: Promise<number[]>;
    private _goodSet: setCB;
    private _badSet: setCB;
    private _latestSet = 0;
    private _tracksLoaded: (updates: TrackUpdate[]) => void;
    private _rt: RT;

    constructor({ userID, goodSet, badSet, tracksLoaded, newTrack }: ControllerArgs) {
        this._user = userID;
        this._sets = this._listSets();
        this._goodSet = goodSet;
        this._badSet = badSet;
        this._tracksLoaded = tracksLoaded;
        this._rt = new RT(userID, (wrapptedTU) => newTrack(wrapptedTU.update));
        window.addEventListener('popstate', (event) => {
            this.selectSet(event.state, false);
        });
    }

    private async _listSets(): Promise<number[]> {
        return fetch(`/_trackUpdate/${this._user}`).then((resp) => resp.json())
            .then((resp: { sessions: number[] }) => {
                let sets = resp.sessions.toSorted().reverse();
                if (sets) {
                    this._latestSet = sets[0];
                }
                return sets;
            });
    }

    getSets(): Promise<number[]> {
        return this._sets;
    }

    selectSet(setID: number, pushState = true) {
        this._sets.then((sets) => {
            if (setID == 0 && sets.length) {
                setID = sets[0]
            }
            if (pushState) {
                history.pushState(setID, setID.toString(), `/u/${this._user}/${setID}`);
            }
            if (!sets.includes(setID)) {
                this._badSet(setID);
                return;
            }
            this._loadSet(setID);
        });
    }

    private _loadSet(setID: number) {
        this._goodSet(setID);
        fetch(`/_trackUpdate/${this._user}/${setID}`).then((resp) => resp.json())
            .then((resp: { updates: TrackUpdate[] }) => {
                if (setID == this._latestSet) {
                    this._rt.connect();
                } else {
                    this._rt.close();
                }
                if (resp.updates && resp.updates.length) {
                    this._tracksLoaded(resp.updates);
                }
            });
    }
}

type WrappedTU = {
    userID: string;
    started: bigint;
    update: TrackUpdate;
};

class RT {
    private _socket: WebSocket;
    private _addr: URL;
    private _newTrack: (tu: WrappedTU) => void;

    constructor(userID: string, newTrack = (tu: WrappedTU) => { }) {
        this._newTrack = newTrack;
        this._addr = new URL(document.location.toString());
        this._addr.protocol = this._addr.protocol == 'https:' ? 'wss' : 'ws';
        this._addr.pathname = `/_sub/${userID}`;
    }

    connect() {
        this._socket = new WebSocket(this._addr.toString());
        this._socket.addEventListener('open', (ev) => this._socketOpened(ev));
        this._socket.addEventListener('close', (ev) => this._socketClosed(ev));
        this._socket.addEventListener('error', (ev) => this._socketError(ev));
        this._socket.addEventListener('message', (ev) => this._socketMessage(ev));
    }

    close() {
        if (!this._socket) {
            return;
        }
        this._socket.close()
    }

    private _socketOpened(ev: Event) {
        console.log('socket opened');
    }
    private _socketClosed(ev: Event) {
        console.log('socket closed');
    }
    private _socketError(ev: Event) {
        console.log('socket error: ', ev);
    }
    private _socketMessage(ev: MessageEvent) {
        let wtu: WrappedTU = JSON.parse(ev.data);
        this._newTrack(wtu);
    }
}

export { Controller, setCB };