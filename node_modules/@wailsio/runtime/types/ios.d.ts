export declare namespace Haptics {
    type ImpactStyle = "light" | "medium" | "heavy" | "soft" | "rigid";
    function Impact(style?: ImpactStyle): Promise<void>;
}
export declare namespace Device {
    interface Info {
        model: string;
        systemName: string;
        systemVersion: string;
        isSimulator: boolean;
    }
    function Info(): Promise<Info>;
}
