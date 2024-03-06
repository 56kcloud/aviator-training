package reservation

import "aviator/errors"

var ReservationAircraftAccessDenyError = errors.AviatorError{
	Id: "reservation_aircraft_access_deny",
	Message: errors.Message{
		EN: "You do not have permission to reserve the selected aircraft",
		FR: "Vous n'avez pas l'autorisation de réserver l'avion sélectionné",
	},
	ApiError: 401,
}

var ReservationRoomAccessDenyError = errors.AviatorError{
	Id: "reservation_aircraft_access_deny",
	Message: errors.Message{
		EN: "You do not have permission to reserve the selected room",
		FR: "Vous n'avez pas l'autorisation de réserver la salle sélectionnée",
	},
	ApiError: 401,
}

var ReservationInvalidBookerError = errors.AviatorError{
	Id: "reservation_invalid_booker",
	Message: errors.Message{
		EN: "The selected booker does not exist",
		FR: "Le reservateur sélectionné n'existe pas",
	},
	ApiError: 400,
}

var ReservationDoubleBookerError = errors.AviatorError{
	Id: "reservation_double_booker",
	Message: errors.Message{
		EN: "You cannot reserve the same aicraft more than once for a given slot",
		FR: "Vous ne pouvez pas réserver le même avion plus d'une fois pour un créneau donné.",
	},
	ApiError: 400,
}

var ReservationInvalidPilotError = errors.AviatorError{
	Id: "reservation_invalid_pilot",
	Message: errors.Message{
		EN: "The selected pilot does not exist",
		FR: "Le pilote sélectionné n'existe pas",
	},
	ApiError: 400,
}

var ReservationExpiredSepRatingError = errors.AviatorError{
	Id: "reservation_invalid_pilot",
	Message: errors.Message{
		EN: "You're SEP rating has expired",
		FR: "Votre qualification SEP est expirée",
	},
	ApiError: 401,
}

var ReservationExpiredMedicalRatingError = errors.AviatorError{
	Id: "reservation_invalid_pilot",
	Message: errors.Message{
		EN: "You're medical  has expired",
		FR: "Votre license médicale est expirée",
	},
	ApiError: 401,
}

var ReservationInvalidAircraftError = errors.AviatorError{
	Id: "reservation_invalid_aircraft",
	Message: errors.Message{
		EN: "The selected aircraft does not exist",
		FR: "L'appareil sélectionné n'existe pas",
	},
	ApiError: 400,
}

var ReservationInvalidReservationTypeError = errors.AviatorError{
	Id: "reservation_invalid_reservation_type",
	Message: errors.Message{
		EN: "The selected reservation type is invalid",
		FR: "Le type de réservation sélectionné n'existe pas",
	},
	ApiError: 400,
}

var ReservationCreateTimePastError = errors.AviatorError{
	Id: "reservation_create_time_in_past",
	Message: errors.Message{
		EN: "The start or end time of a reservation cannot be in the past",
		FR: "L'heure de début ou de fin d'une réservation ne peut pas se situer dans le passé",
	},
	ApiError: 400,
}

var ReservationTimesSwappedError = errors.AviatorError{
	Id: "reservation_times_swapped",
	Message: errors.Message{
		EN: "The start time of a reservation must be before the end time",
		FR: "L'heure de début d'une réservation doit être antérieure à l'heure de fin",
	},
	ApiError: 400,
}

var ReservationTimesEqualError = errors.AviatorError{
	Id: "reservation_times_equal",
	Message: errors.Message{
		EN: "The start and end time must be different",
		FR: "Les heures de début et de fin doivent être différentes",
	},
	ApiError: 400,
}

var ReservationInvalidInstructorError = errors.AviatorError{
	Id: "reservation_invalid_intructor",
	Message: errors.Message{
		EN: "The selected instructor is invalid",
		FR: "L'instructeur sélectionné n'est pas valide",
	},
	ApiError: 400,
}

var ReservationInstructorRequiredError = errors.AviatorError{
	Id: "reservation_instructor_required",
	Message: errors.Message{
		EN: "The selected reservation type requires an instructor",
		FR: "Le type de réservation sélectionné nécessite un instructeur",
	},
	ApiError: 400,
}

var ReservationPriorityLostError = errors.AviatorError{
	Id: "reservation_priority_lost",
	Message: errors.Message{
		EN: "Updating this reservation would cause it to lose its priority",
		FR: "La mise à jour de cette réservation lui ferait perdre sa priorité",
	},
	ApiError: 400,
}

var ReservationOverbookingConflictError = errors.AviatorError{
	Id: "reservation_overbooking_conflict",
	Message: errors.Message{
		EN: "Time slot cannot be reserved due to existing reservations",
		FR: "Le créneau horaire ne peut être réservé en raison de réservations existantes",
	},
	ApiError: 400,
}

var ReservationPastUpdateError = errors.AviatorError{
	Id: "reservation_past_update",
	Message: errors.Message{
		EN: "A reservation in the past cannot be updated",
		FR: "Une réservation dans le passé ne peut pas être mise à jour",
	},
	ApiError: 400,
}

var ReservationUnauthorizedError = errors.AviatorError{
	Id: "reservation_unauthorized",
	Message: errors.Message{
		EN: "You do not have permission to perform this action",
		FR: "Vous n'êtes pas autorisé à effectuer cette action",
	},
	ApiError: 401,
}

var ReservationTimeRangeError = errors.AviatorError{
	Id: "reservation_time_range",
	Message: errors.Message{
		EN: "You must provide both a start and end date",
		FR: "Vous devez indiquer une date de début et une date de fin",
	},
	ApiError: 401,
}

var ReservationSamePilotInstructorError = errors.AviatorError{
	Id: "reservation_same_pilot_instructor",
	Message: errors.Message{
		EN: "The pilot and instructor cannot be the same",
		FR: "Le pilote et l'instructeur ne peuvent pas être les mêmes",
	},
	ApiError: 401,
}
