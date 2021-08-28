## Federation Interconnect

#### Operator sets up Cloudlets & Availability Zones (AZ)

| User Action                                          | OP1                                     | OP2                                     |
| ---------------------------------------------------- | --------------------------------------- | --------------------------------------- |
| Operator1 onboards cloudlets                         | - cloudlet onboarding                   |                                         |
| Operator1 creates AZs to be shared during federation | - create zone = collection of cloudlets |                                         |
| Operator2 onboards cloudlets                         |                                         | - cloudlet onboarding                   |
| Operator2 creates AZs to be shared during federation |                                         | - create zone = collection of cloudlets |

#### Operator sets up Federation and adds another OP as partner

| User Action                         | OP1                                       | OP2                                       |
| ----------------------------------- | ----------------------------------------- | ----------------------------------------- |
| Operator1 sets up federation        | - create new federation ID                |                                           |
|                                     | - store federation details in DB          |                                           |
| Operator2 sets up federation        |                                           | - create new federation ID                |
|                                     |                                           | - store federation details i DB           |
| Operator1 adds Operator2 as partner | - calls POST `/operator/partner`          |                                           |
|                                     |                                           | - validate federation                     |
|                                     |                                           | - stores partner federation details in DB |
|                                     |                                           | - shares all the zones                    |
|                                     | - receives all the zones                  |                                           |
|                                     | - stores partner federation details in DB |                                           |
|                                     | - stores zone info in DB                  |                                           |

#### Operator adds new zone after federation is setup

| User Action | OP1  | OP2  |
| ----------- | ---- | ---- |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |
|             |      |      |

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
|                                                     |                                                              | -                                                            |
|                                                     |                                                              |                                                              |

