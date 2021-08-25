## Federation Interconnect

#### Operator sets up Cloudlets & Availability Zones (AZ)

| User Action                                          | OP1                                     | OP2                                     |
| ---------------------------------------------------- | --------------------------------------- | --------------------------------------- |
| Operator1 onboards cloudlets                         | - cloudlet onboarding                   |                                         |
| Operator1 creates AZs to be shared during federation | - create zone = collection of cloudlets |                                         |
| Operator2 onboards cloudlets                         |                                         | - cloudlet onboarding                   |
| Operator2 creates AZs to be shared during federation |                                         | - create zone = collection of cloudlets |

#### OP1 sets up Federation and adds OP2 as partner

| User Action                         | OP1                                       | OP2                             |
| ----------------------------------- | ----------------------------------------- | ------------------------------- |
| Operator1 sets up federation        | - create new federation ID                |                                 |
|                                     | - store federation details in DB          |                                 |
| Operator2 sets up federation        |                                           | - create new federation ID      |
|                                     |                                           | - store federation details i DB |
| Operator1 adds Operator2 as partner | - calls POST `/operator/partner`          |                                 |
|                                     |                                           | - validate federation           |
|                                     |                                           | - stores partner details in DB  |
|                                     |                                           | - shares all the zones          |
|                                     | - receives all the zones                  |                                 |
|                                     | - stores partner federation details in DB |                                 |
|                                     | - stores zone info in DB                  |                                 |

