## Federation Interconnect

#### Operator sets up Cloudlets & Availability Zones (AZ)

| User Action                                          | OP1                                     | OP2                                     |
| ---------------------------------------------------- | --------------------------------------- | --------------------------------------- |
| Operator1 onboards cloudlets                         | - cloudlet onboarding                   |                                         |
| Operator1 creates AZs to be shared during federation | - create zone = collection of cloudlets |                                         |
| Operator2 onboards cloudlets                         |                                         | - cloudlet onboarding                   |
| Operator2 creates AZs to be shared during federation |                                         | - create zone = collection of cloudlets |

#### Operator sets up Federation and adds another OP as partner

| User Action                         | OP1                                                          | OP2                                       |
| ----------------------------------- | ------------------------------------------------------------ | ----------------------------------------- |
| Operator1 sets up federation        | - create new federation ID                                   |                                           |
|                                     | - store federation details in DB                             |                                           |
| Operator2 sets up federation        |                                                              | - create new federation ID                |
|                                     |                                                              | - store federation details in DB          |
| Operator1 adds Operator2 as partner | - calls **POST** `/operator/partner`                         |                                           |
|                                     |                                                              | - validate federation                     |
|                                     |                                                              | - stores partner federation details in DB |
|                                     |                                                              | - shares all the zones                    |
|                                     | - receives all the zones                                     |                                           |
|                                     | - stores partner federation details in DB, set the partner OP role as **access** (since OP1 can access its zones) |                                           |
|                                     | - stores zone info in DB                                     |                                           |

#### Operator updates its Federation attributes

| User Action                                                  | OP1                                         | OP2                                                          |
| ------------------------------------------------------------ | ------------------------------------------- | ------------------------------------------------------------ |
| Operator2 updates its federation attributes like MCC/MNC/LocatorEndPoint |                                             | - validate update fields                                     |
|                                                              |                                             | - store updated details in DB                                |
|                                                              |                                             | - calls **PUT** `/operator/partner` for all the partner OPs (with whom this OP is sharing zones i.e. role is to **share** zones) about the update of federation attributes |
|                                                              | - validate federation                       |                                                              |
|                                                              | - update federation attributes of OP2 in DB |                                                              |

#### Operator adds new zone after federation is setup (share zone)

| User Action                | OP1                        | OP2                                                          |
| -------------------------- | -------------------------- | ------------------------------------------------------------ |
| Operator2 creates new zone |                            | - create new zone                                            |
|                            |                            | - adds group of cloudlets under this zone                    |
|                            |                            | - store zone in DB                                           |
|                            |                            | - calls **POST** `/operator/notify/zone` to notify partner OP (with whom this OP is sharing zones i.e. role is to **share** zones) about this zone |
|                            | - validate federation      |                                                              |
|                            | - store zone details in DB |                                                              |

#### Operator deletes a zone after federation is setup (unshare zone)

| User Action               | OP1                           | OP2                                                          |
| ------------------------- | ----------------------------- | ------------------------------------------------------------ |
| Operator2 deletes a  zone |                               | - delete zone                                                |
|                           |                               | - ensure that it is not being used/registered by an OP       |
|                           |                               | - store zone in DB                                           |
|                           |                               | - calls **DELETE** `/operator/notify/zone` to notify partner OP (with whom this OP is sharing zones i.e. role is to **share** zones) about the deletion of this zone |
|                           | - validate federation         |                                                              |
|                           | - delete zone details from DB |                                                              |

#### Operator registers a zone after federation is setup

| User Action                   | OP1                                | OP2                                            |
| ----------------------------- | ---------------------------------- | ---------------------------------------------- |
| Operator1 registers OP2  zone | - check if zone exists             |                                                |
|                               | - calls **POST** `/operator/zone`  |                                                |
|                               |                                    | - validate federation                          |
|                               |                                    | - store registering OP details along with zone |
|                               | - store OP details along with zone |                                                |

#### Operator deregisters a zone after federation is setup

| User Action                     | OP1                                 | OP2                            |
| ------------------------------- | ----------------------------------- | ------------------------------ |
| Operator1 deregisters OP2  zone | - check if zone exists              |                                |
|                                 | - calls **DELETE** `/operator/zone` |                                |
|                                 |                                     | - validate federation          |
|                                 |                                     | - delete registered OP details |
|                                 | - delete registered OP details      |                                |

#### Operator removes partner federation

| User Action                                | OP1                                                       | OP2                   |
| ------------------------------------------ | --------------------------------------------------------- | --------------------- |
| Operator1 remove OP2 as federation partner | - validate that all the partner OP zones are deregistered |                       |
|                                            | - calls **DELETE** `/operator/partner`                    |                       |
|                                            |                                                           | - validate federation |
|                                            |                                                           | - delete OP1 details  |
|                                            | - delete all details of OP2 zones                         |                       |
|                                            | - delete OP2 federation details                           |                       |

#### Developer onboards application on a federation zone

| User Action                                         | OP1                                                          | OP2                                                          |
| --------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| Developer onboards application on OP2 zone from OP1 | - validate application data                                  |                                                              |
|                                                     | - mark onboarding status as `uploadpending` for this zone in the DB |                                                              |
|                                                     | - upload artifact on partner edge, calls POST  `/operator/artifacts` |                                                              |
|                                                     |                                                              | - validate federation                                        |
|                                                     |                                                              | - upload artifact to OP's registry                           |
|                                                     |                                                              | - notify caller about upload status, calls GET  `/operator/artifact/status` |
|                                                     | - mark onboarding status as `uploaded` for this zone in the DB |                                                              |
|                                                     | - onboard app, calls POST `/operator/application/onboarding` |                                                              |
|                                                     |                                                              | - validate federation                                        |

