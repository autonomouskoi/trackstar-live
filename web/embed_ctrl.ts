import { bus, enumName } from "/bus.js";

import * as livepb from "/m/trackstar-live/pb/trackstar-live/live_pb.js";
import { Cfg } from './controller.js';
import { Tokens } from './token.js';
import { Status } from "./status.js";

const TOPIC_REQUEST = enumName(livepb.BusTopic, livepb.BusTopic.TRACKSTAR_LIVE_REQUEST);

function start(mainContainer: HTMLElement) {
    let cfg = new Cfg();

    mainContainer.classList.add('flex-column');
    mainContainer.style.setProperty('gap', '1rem');

    bus.waitForTopic(TOPIC_REQUEST, 5000)
        .then(() => {
            mainContainer.appendChild(new Status());
            mainContainer.appendChild(new Tokens(cfg));
            cfg.refresh();
        });
}

export { start };