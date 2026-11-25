package wf2

// Packet sizes and field offsets follow JonathanSoszka/wf2-telemetry layout.json
// (MIT), cross-checked against vazhure/vAzhureRacingHub Wreckfest2Structs.cs.

const (
	Signature = "pkro"

	ParticipantsMax = 36

	HeaderSize = 14

	PacketMainSize                    = 1218
	PacketParticipantsLeaderboardSize = 1231
	PacketParticipantsTimingSize      = 1231
	PacketParticipantsSectorsSize     = 1375
	PacketParticipantsMotionSize      = 1383
	PacketParticipantsInfoSize        = 8195
	PacketParticipantsDamageSize      = 1027

	StrideLeaderboard = 32
	StrideTiming      = 32
	StrideSectors     = 36
	StrideMotion      = 38
	StrideInfo        = 213
	StrideDamage      = 21

	DamagePartsMax       = 0x38 // 56 packed slots (3 bits each → 21 bytes)
	DamageBitsPerPart    = 3
	DamageBytesPerDriver = 21
	DamagePartCount      = 0x25 // 37 named parts in DamagePart enum

	TireCount = 4
	TireSize  = 72

	MainMarshalFlags = 14
	MainPlayerBoard  = 16
	MainPlayerTiming = 48
	MainPlayerSector = 80
	MainPlayerInfo   = 116
	MainPlayerDamage = 329
	MainCarPlayer    = 350
	MainSession      = 882
	MainSessionEnd   = 1090
	MainPlayerStatus = 1090
	MainInputExt     = 1092

	CarAssistsOff   = 0
	CarChassisOff   = 8
	CarDrivelineOff = 60
	CarEngineOff    = 84
	CarInputOff     = 140
	CarOrientOff    = 160
	CarVelocityOff  = 194
	CarTiresOff     = 230
)

// PacketSizes indexes PacketType → expected UDP datagram length.
var PacketSizes = [...]int{
	PacketMainSize,
	PacketParticipantsLeaderboardSize,
	PacketParticipantsTimingSize,
	PacketParticipantsSectorsSize,
	PacketParticipantsMotionSize,
	PacketParticipantsInfoSize,
	PacketParticipantsDamageSize,
}
