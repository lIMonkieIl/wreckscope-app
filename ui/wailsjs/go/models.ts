export namespace racefeed {
	
	export class CarAssists {
	    flags: number;
	    assistGearbox: number;
	    levelAbs: number;
	    levelTcs: number;
	    levelEsc: number;
	
	    static createFrom(source: any = {}) {
	        return new CarAssists(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flags = source["flags"];
	        this.assistGearbox = source["assistGearbox"];
	        this.levelAbs = source["levelAbs"];
	        this.levelTcs = source["levelTcs"];
	        this.levelEsc = source["levelEsc"];
	    }
	}
	export class CarChassis {
	    trackWidth: number[];
	    wheelBase: number;
	    steeringWheelLockToLock: number;
	    steeringLock: number;
	    cornerWeights: number[];
	
	    static createFrom(source: any = {}) {
	        return new CarChassis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trackWidth = source["trackWidth"];
	        this.wheelBase = source["wheelBase"];
	        this.steeringWheelLockToLock = source["steeringWheelLockToLock"];
	        this.steeringLock = source["steeringLock"];
	        this.cornerWeights = source["cornerWeights"];
	    }
	}
	export class CarDriveline {
	    type: number;
	    gear: number;
	    gearMax: number;
	    speed: number;
	
	    static createFrom(source: any = {}) {
	        return new CarDriveline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.gear = source["gear"];
	        this.gearMax = source["gearMax"];
	        this.speed = source["speed"];
	    }
	}
	export class CarEngine {
	    flags: number;
	    rpm: number;
	    rpmMax: number;
	    rpmRedline: number;
	    rpmIdle: number;
	    torque: number;
	    power: number;
	    tempBlock: number;
	    tempWater: number;
	    pressureManifold: number;
	    pressureOil: number;
	
	    static createFrom(source: any = {}) {
	        return new CarEngine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flags = source["flags"];
	        this.rpm = source["rpm"];
	        this.rpmMax = source["rpmMax"];
	        this.rpmRedline = source["rpmRedline"];
	        this.rpmIdle = source["rpmIdle"];
	        this.torque = source["torque"];
	        this.power = source["power"];
	        this.tempBlock = source["tempBlock"];
	        this.tempWater = source["tempWater"];
	        this.pressureManifold = source["pressureManifold"];
	        this.pressureOil = source["pressureOil"];
	    }
	}
	export class CarTire {
	    rps: number;
	    camber: number;
	    slipRatio: number;
	    slipAngle: number;
	    radiusUnloaded: number;
	    loadVertical: number;
	    forceLat: number;
	    forceLong: number;
	    temperatureInner: number;
	    temperatureTread: number;
	    suspensionVelocity: number;
	    suspensionDisplacement: number;
	    suspensionDispNorm: number;
	    positionVertical: number;
	    surfaceType: number;
	
	    static createFrom(source: any = {}) {
	        return new CarTire(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rps = source["rps"];
	        this.camber = source["camber"];
	        this.slipRatio = source["slipRatio"];
	        this.slipAngle = source["slipAngle"];
	        this.radiusUnloaded = source["radiusUnloaded"];
	        this.loadVertical = source["loadVertical"];
	        this.forceLat = source["forceLat"];
	        this.forceLong = source["forceLong"];
	        this.temperatureInner = source["temperatureInner"];
	        this.temperatureTread = source["temperatureTread"];
	        this.suspensionVelocity = source["suspensionVelocity"];
	        this.suspensionDisplacement = source["suspensionDisplacement"];
	        this.suspensionDispNorm = source["suspensionDispNorm"];
	        this.positionVertical = source["positionVertical"];
	        this.surfaceType = source["surfaceType"];
	    }
	}
	export class CarMotionVelocity {
	    velocityLocalX: number;
	    velocityLocalY: number;
	    velocityLocalZ: number;
	    angularVelocityX: number;
	    angularVelocityY: number;
	    angularVelocityZ: number;
	    accelerationLocalX: number;
	    accelerationLocalY: number;
	    accelerationLocalZ: number;
	
	    static createFrom(source: any = {}) {
	        return new CarMotionVelocity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.velocityLocalX = source["velocityLocalX"];
	        this.velocityLocalY = source["velocityLocalY"];
	        this.velocityLocalZ = source["velocityLocalZ"];
	        this.angularVelocityX = source["angularVelocityX"];
	        this.angularVelocityY = source["angularVelocityY"];
	        this.angularVelocityZ = source["angularVelocityZ"];
	        this.accelerationLocalX = source["accelerationLocalX"];
	        this.accelerationLocalY = source["accelerationLocalY"];
	        this.accelerationLocalZ = source["accelerationLocalZ"];
	    }
	}
	export class CarMotionOrientation {
	    positionX: number;
	    positionY: number;
	    positionZ: number;
	    orientationQuaternionX: number;
	    orientationQuaternionY: number;
	    orientationQuaternionZ: number;
	    orientationQuaternionW: number;
	    extentsX: number;
	    extentsY: number;
	    extentsZ: number;
	
	    static createFrom(source: any = {}) {
	        return new CarMotionOrientation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.positionX = source["positionX"];
	        this.positionY = source["positionY"];
	        this.positionZ = source["positionZ"];
	        this.orientationQuaternionX = source["orientationQuaternionX"];
	        this.orientationQuaternionY = source["orientationQuaternionY"];
	        this.orientationQuaternionZ = source["orientationQuaternionZ"];
	        this.orientationQuaternionW = source["orientationQuaternionW"];
	        this.extentsX = source["extentsX"];
	        this.extentsY = source["extentsY"];
	        this.extentsZ = source["extentsZ"];
	    }
	}
	export class CarInput {
	    throttle: number;
	    brake: number;
	    clutch: number;
	    handbrake: number;
	    steering: number;
	
	    static createFrom(source: any = {}) {
	        return new CarInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.throttle = source["throttle"];
	        this.brake = source["brake"];
	        this.clutch = source["clutch"];
	        this.handbrake = source["handbrake"];
	        this.steering = source["steering"];
	    }
	}
	export class CarFull {
	    assists: CarAssists;
	    chassis: CarChassis;
	    driveline: CarDriveline;
	    engine: CarEngine;
	    input: CarInput;
	    orientation: CarMotionOrientation;
	    velocity: CarMotionVelocity;
	    tires: CarTire[];
	
	    static createFrom(source: any = {}) {
	        return new CarFull(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.assists = this.convertValues(source["assists"], CarAssists);
	        this.chassis = this.convertValues(source["chassis"], CarChassis);
	        this.driveline = this.convertValues(source["driveline"], CarDriveline);
	        this.engine = this.convertValues(source["engine"], CarEngine);
	        this.input = this.convertValues(source["input"], CarInput);
	        this.orientation = this.convertValues(source["orientation"], CarMotionOrientation);
	        this.velocity = this.convertValues(source["velocity"], CarMotionVelocity);
	        this.tires = this.convertValues(source["tires"], CarTire);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class InputExtended {
	    flags: number;
	    ffbForce: number;
	
	    static createFrom(source: any = {}) {
	        return new InputExtended(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flags = source["flags"];
	        this.ffbForce = source["ffbForce"];
	    }
	}
	export class ParticipantDamage {
	    damageStates: number[];
	    parts?: number[];
	
	    static createFrom(source: any = {}) {
	        return new ParticipantDamage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.damageStates = source["damageStates"];
	        this.parts = source["parts"];
	    }
	}
	export class ParticipantInfo {
	    carId: string;
	    carName: string;
	    playerName: string;
	    participantIndex: number;
	    lastNormalTrackStatusTime: number;
	    lastCollisionTime: number;
	    lastResetTime: number;
	
	    static createFrom(source: any = {}) {
	        return new ParticipantInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.carId = source["carId"];
	        this.carName = source["carName"];
	        this.playerName = source["playerName"];
	        this.participantIndex = source["participantIndex"];
	        this.lastNormalTrackStatusTime = source["lastNormalTrackStatusTime"];
	        this.lastCollisionTime = source["lastCollisionTime"];
	        this.lastResetTime = source["lastResetTime"];
	    }
	}
	export class ParticipantLeaderboard {
	    status: number;
	    trackStatus: number;
	    lapCurrent: number;
	    position: number;
	    health: number;
	    wrecks: number;
	    frags: number;
	    assists: number;
	    score: number;
	    points: number;
	    deltaLeader: number;
	    lapTiming: number;
	
	    static createFrom(source: any = {}) {
	        return new ParticipantLeaderboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.trackStatus = source["trackStatus"];
	        this.lapCurrent = source["lapCurrent"];
	        this.position = source["position"];
	        this.health = source["health"];
	        this.wrecks = source["wrecks"];
	        this.frags = source["frags"];
	        this.assists = source["assists"];
	        this.score = source["score"];
	        this.points = source["points"];
	        this.deltaLeader = source["deltaLeader"];
	        this.lapTiming = source["lapTiming"];
	    }
	}
	export class ParticipantTiming {
	    lapTimeCurrent: number;
	    lapTimePenaltyCurrent: number;
	    lapTimeLast: number;
	    lapTimeBest: number;
	    lapBest: number;
	    deltaAhead: number;
	    deltaBehind: number;
	    lapProgress: number;
	
	    static createFrom(source: any = {}) {
	        return new ParticipantTiming(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lapTimeCurrent = source["lapTimeCurrent"];
	        this.lapTimePenaltyCurrent = source["lapTimePenaltyCurrent"];
	        this.lapTimeLast = source["lapTimeLast"];
	        this.lapTimeBest = source["lapTimeBest"];
	        this.lapBest = source["lapBest"];
	        this.deltaAhead = source["deltaAhead"];
	        this.deltaBehind = source["deltaBehind"];
	        this.lapProgress = source["lapProgress"];
	    }
	}
	export class ParticipantTimingSectors {
	    sectorTimeCurrentLap1: number;
	    sectorTimeCurrentLap2: number;
	    sectorTimeLastLap1: number;
	    sectorTimeLastLap2: number;
	    sectorTimeBestLap1: number;
	    sectorTimeBestLap2: number;
	    sectorTimeBest1: number;
	    sectorTimeBest2: number;
	    sectorTimeBest3: number;
	
	    static createFrom(source: any = {}) {
	        return new ParticipantTimingSectors(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sectorTimeCurrentLap1 = source["sectorTimeCurrentLap1"];
	        this.sectorTimeCurrentLap2 = source["sectorTimeCurrentLap2"];
	        this.sectorTimeLastLap1 = source["sectorTimeLastLap1"];
	        this.sectorTimeLastLap2 = source["sectorTimeLastLap2"];
	        this.sectorTimeBestLap1 = source["sectorTimeBestLap1"];
	        this.sectorTimeBestLap2 = source["sectorTimeBestLap2"];
	        this.sectorTimeBest1 = source["sectorTimeBest1"];
	        this.sectorTimeBest2 = source["sectorTimeBest2"];
	        this.sectorTimeBest3 = source["sectorTimeBest3"];
	    }
	}
	export class PlayerTelemetry {
	    marshalFlagsPlayer: number;
	    participantPlayerLeaderboard: ParticipantLeaderboard;
	    participantPlayerTiming: ParticipantTiming;
	    participantPlayerTimingSectors: ParticipantTimingSectors;
	    participantPlayerInfo: ParticipantInfo;
	    participantPlayerDamage: ParticipantDamage;
	    carPlayer: CarFull;
	    playerStatusFlags: number;
	    inputExtended: InputExtended;
	
	    static createFrom(source: any = {}) {
	        return new PlayerTelemetry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marshalFlagsPlayer = source["marshalFlagsPlayer"];
	        this.participantPlayerLeaderboard = this.convertValues(source["participantPlayerLeaderboard"], ParticipantLeaderboard);
	        this.participantPlayerTiming = this.convertValues(source["participantPlayerTiming"], ParticipantTiming);
	        this.participantPlayerTimingSectors = this.convertValues(source["participantPlayerTimingSectors"], ParticipantTimingSectors);
	        this.participantPlayerInfo = this.convertValues(source["participantPlayerInfo"], ParticipantInfo);
	        this.participantPlayerDamage = this.convertValues(source["participantPlayerDamage"], ParticipantDamage);
	        this.carPlayer = this.convertValues(source["carPlayer"], CarFull);
	        this.playerStatusFlags = source["playerStatusFlags"];
	        this.inputExtended = this.convertValues(source["inputExtended"], InputExtended);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RaceDetails {
	    trackId: string;
	    track: string;
	    trackName?: string;
	    distance?: number;
	    trackLength?: number;
	    laps: number;
	    eventLength?: number;
	    mode: number;
	    gameMode?: number;
	    damage: number;
	    damageMode?: number;
	    grid: number;
	    gridSize?: number;
	    remaining: number;
	    gridSizeRemaining?: number;
	    sectorCount?: number;
	    sectorFract1?: number;
	    sectorFract2?: number;
	    sessionStatus?: number;
	    status?: number;
	    flags: number;
	    gameStatus?: number;
	    sessionMs?: number;
	    clock: number;
	
	    static createFrom(source: any = {}) {
	        return new RaceDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trackId = source["trackId"];
	        this.track = source["track"];
	        this.trackName = source["trackName"];
	        this.distance = source["distance"];
	        this.trackLength = source["trackLength"];
	        this.laps = source["laps"];
	        this.eventLength = source["eventLength"];
	        this.mode = source["mode"];
	        this.gameMode = source["gameMode"];
	        this.damage = source["damage"];
	        this.damageMode = source["damageMode"];
	        this.grid = source["grid"];
	        this.gridSize = source["gridSize"];
	        this.remaining = source["remaining"];
	        this.gridSizeRemaining = source["gridSizeRemaining"];
	        this.sectorCount = source["sectorCount"];
	        this.sectorFract1 = source["sectorFract1"];
	        this.sectorFract2 = source["sectorFract2"];
	        this.sessionStatus = source["sessionStatus"];
	        this.status = source["status"];
	        this.flags = source["flags"];
	        this.gameStatus = source["gameStatus"];
	        this.sessionMs = source["sessionMs"];
	        this.clock = source["clock"];
	    }
	}
	export class Settings {
	    drawTrackOutline: boolean;
	    cautionSeconds: number;
	    alertSeconds: number;
	    speedThreshold: number;
	    hideBots: boolean;
	    speedUnit: string;
	    muteAlertSound: boolean;
	    theme: string;
	    mode: string;
	    view: string;
	    telemetryConfigPath?: string;
	    setupCompleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.drawTrackOutline = source["drawTrackOutline"];
	        this.cautionSeconds = source["cautionSeconds"];
	        this.alertSeconds = source["alertSeconds"];
	        this.speedThreshold = source["speedThreshold"];
	        this.hideBots = source["hideBots"];
	        this.speedUnit = source["speedUnit"];
	        this.muteAlertSound = source["muteAlertSound"];
	        this.theme = source["theme"];
	        this.mode = source["mode"];
	        this.view = source["view"];
	        this.telemetryConfigPath = source["telemetryConfigPath"];
	        this.setupCompleted = source["setupCompleted"];
	    }
	}
	export class Standing {
	    trackStatus: number;
	    frags: number;
	    wrecks: number;
	    assists?: number;
	    score?: number;
	    points?: number;
	    lapTiming?: number;
	    leaderboard?: ParticipantLeaderboard;
	    timing?: ParticipantTiming;
	    sectors?: ParticipantTimingSectors;
	    info?: ParticipantInfo;
	    damage?: ParticipantDamage;
	    progress?: number;
	    bestSectors?: number[];
	    slot: number;
	    name: string;
	    carId: string;
	    carName?: string;
	    status: number;
	    place: number;
	    lap: number;
	    health: number;
	    gap?: number;
	    current?: number;
	    last?: number;
	    best?: number;
	
	    static createFrom(source: any = {}) {
	        return new Standing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trackStatus = source["trackStatus"];
	        this.frags = source["frags"];
	        this.wrecks = source["wrecks"];
	        this.assists = source["assists"];
	        this.score = source["score"];
	        this.points = source["points"];
	        this.lapTiming = source["lapTiming"];
	        this.leaderboard = this.convertValues(source["leaderboard"], ParticipantLeaderboard);
	        this.timing = this.convertValues(source["timing"], ParticipantTiming);
	        this.sectors = this.convertValues(source["sectors"], ParticipantTimingSectors);
	        this.info = this.convertValues(source["info"], ParticipantInfo);
	        this.damage = this.convertValues(source["damage"], ParticipantDamage);
	        this.progress = source["progress"];
	        this.bestSectors = source["bestSectors"];
	        this.slot = source["slot"];
	        this.name = source["name"];
	        this.carId = source["carId"];
	        this.carName = source["carName"];
	        this.status = source["status"];
	        this.place = source["place"];
	        this.lap = source["lap"];
	        this.health = source["health"];
	        this.gap = source["gap"];
	        this.current = source["current"];
	        this.last = source["last"];
	        this.best = source["best"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Snapshot {
	    race?: RaceDetails;
	    player?: PlayerTelemetry;
	    standings: Standing[];
	    mode: string;
	    connection: string;
	    lastPacket: string;
	    packetCount: number;
	    packetTypes: Record<number, number>;
	    cars: recovery.Car[];
	    settings: Settings;
	    sessionRacing: boolean;
	    raceEpoch: number;
	    localSlot?: number;
	    medics: number[];
	    participantVisibility?: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.race = this.convertValues(source["race"], RaceDetails);
	        this.player = this.convertValues(source["player"], PlayerTelemetry);
	        this.standings = this.convertValues(source["standings"], Standing);
	        this.mode = source["mode"];
	        this.connection = source["connection"];
	        this.lastPacket = source["lastPacket"];
	        this.packetCount = source["packetCount"];
	        this.packetTypes = source["packetTypes"];
	        this.cars = this.convertValues(source["cars"], recovery.Car);
	        this.settings = this.convertValues(source["settings"], Settings);
	        this.sessionRacing = source["sessionRacing"];
	        this.raceEpoch = source["raceEpoch"];
	        this.localSlot = source["localSlot"];
	        this.medics = source["medics"];
	        this.participantVisibility = source["participantVisibility"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace recovery {
	
	export class ScoreFactor {
	    label: string;
	    points: number;
	
	    static createFrom(source: any = {}) {
	        return new ScoreFactor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.points = source["points"];
	    }
	}
	export class Car {
	    extents?: number[];
	    orientation?: number[];
	    scoreFactors?: ScoreFactor[];
	    slot: number;
	    name: string;
	    carId?: string;
	    x: number;
	    y: number;
	    z: number;
	    speed: number;
	    place: number;
	    lap: number;
	    health: number;
	    state: string;
	    stopped: number;
	    reason: string;
	    confidence: string;
	    score: number;
	    // Go type: time
	    updated: any;
	    active: boolean;
	    isBot: boolean;
	    status: number;
	    tilt: number;
	    inverted: boolean;
	    incidentId?: string;
	    // Go type: time
	    incidentSince?: any;
	
	    static createFrom(source: any = {}) {
	        return new Car(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.extents = source["extents"];
	        this.orientation = source["orientation"];
	        this.scoreFactors = this.convertValues(source["scoreFactors"], ScoreFactor);
	        this.slot = source["slot"];
	        this.name = source["name"];
	        this.carId = source["carId"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.z = source["z"];
	        this.speed = source["speed"];
	        this.place = source["place"];
	        this.lap = source["lap"];
	        this.health = source["health"];
	        this.state = source["state"];
	        this.stopped = source["stopped"];
	        this.reason = source["reason"];
	        this.confidence = source["confidence"];
	        this.score = source["score"];
	        this.updated = this.convertValues(source["updated"], null);
	        this.active = source["active"];
	        this.isBot = source["isBot"];
	        this.status = source["status"];
	        this.tilt = source["tilt"];
	        this.inverted = source["inverted"];
	        this.incidentId = source["incidentId"];
	        this.incidentSince = this.convertValues(source["incidentSince"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

