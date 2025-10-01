import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as livepb from "/m/trackstar-live/pb/trackstar-live/live_pb.js";
import { ValueUpdater } from "/vu.js";

const TOPIC_REQUEST = enumName(livepb.BusTopic, livepb.BusTopic.TRACKSTAR_LIVE_REQUEST);
const TOPIC_COMMAND = enumName(livepb.BusTopic, livepb.BusTopic.TRACKSTAR_LIVE_COMMAND);

class Cfg extends ValueUpdater<livepb.Config> {
    constructor() {
        super(new livepb.Config());
    }

    refresh() {
        bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_REQUEST,
            type: livepb.MessageTypeRequest.CONFIG_GET_REQ,
            message: new livepb.GetConfigRequest().toBinary(),
        })).then((reply) => {
            let cgResp = livepb.GetConfigResponse.fromBinary(reply.message);
            this.update(cgResp.config);
        })
    }

    async save(cfg: livepb.Config) {
        let csr = new livepb.SetConfigRequest({
            config: cfg,
        });
        return bus.sendAnd(new buspb.BusMessage({
            topic: TOPIC_COMMAND,
            type: livepb.MessageTypeCommand.CONFIG_SET_REQ,
            message: csr.toBinary(),
        })).then((reply) => {
            let csResp = livepb.SetConfigResponse.fromBinary(reply.message);
            this.update(csResp.config);
        })
    }
}
export { Cfg };