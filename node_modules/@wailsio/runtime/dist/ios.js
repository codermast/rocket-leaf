/*
 _     __     _ __
| |  / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/
import { newRuntimeCaller, objectNames } from "./runtime.js";
const call = newRuntimeCaller(objectNames.IOS);
// Method IDs
const HapticsImpact = 0;
const DeviceInfo = 1;
export var Haptics;
(function (Haptics) {
    function Impact(style = "medium") {
        return call(HapticsImpact, { style });
    }
    Haptics.Impact = Impact;
})(Haptics || (Haptics = {}));
export var Device;
(function (Device) {
    function Info() {
        return call(DeviceInfo);
    }
    Device.Info = Info;
})(Device || (Device = {}));
