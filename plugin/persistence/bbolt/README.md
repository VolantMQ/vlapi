### Persistence BoltDB

[![Build Status](https://travis-ci.org/VolantMQ/persistence-boltdb.svg?branch=master)](https://travis-ci.org/VolantMQ/persistence-boltdb)

BoltDB based persistence plugin for VolantMQ

```
╔════════════════════════════════════╗                     
║ V: count - total amount of sessions║                     
║ B: sessions - actuall sessions     ║                     
╚═▲══════════════════════════════════╝                     
  │                                                        
  │.n                                                      
╔═╩═══════════════╗      .1 ┌─────────────────────┐        
║ V: subscriptions◀─────────┤Encoded subscriptions│        
║ B: state        ◀────┐    └─────────────────────┘        
║ B: packets      ◀───┐│ .1 ╔═════════════╗                
╚═════════════════╝   │└────╣ V: version  ║                
                      │     ║ V: since    ║                
                      │     ║ V: expireIn ║                
                      │     ║ V: willIn   ║                
                      │     ║ V: willData ║                
┌───────────────┐     │     ╚═════════════╝                
│ Legend        │     │  .1 ╔═════╗.1  .n ╔═══════════════╗
│ V - value     │     └─────╣ Seq ◀───────╣ V: data       ║
│ B - bucket    │           ╚═════╝       ║ V: unAck      ║
└───────────────┘                         ║ V: expireAt   ║
                                          ╚═══════════════╝
```